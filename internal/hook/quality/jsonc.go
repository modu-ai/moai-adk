package quality

// jsonc.go — the smallest JSONC-to-JSON reduction the typecheck axis needs.
//
// tsconfig.json is JSONC: TypeScript accepts // and /* */ comments and trailing
// commas, and real configs use both. encoding/json rejects them, which matters
// here because isSolutionStyleTsconfig answers false on a parse failure — an
// unstripped comment would route a solution-style config to "run tsc" and
// reinstate the vacuous pass the check exists to prevent.
//
// This is deliberately not a full JSONC parser. The file is read to answer one
// structural question (files: [] plus references, nothing to compile), so a
// dependency would be more than the question is worth.

// stripJSONC removes comments and trailing commas, leaving valid JSON.
//
// String state is tracked so a "//" inside a value — a URL in a path mapping,
// say — survives, and backslash escapes are honoured so an escaped quote does
// not end a string early.
func stripJSONC(data []byte) []byte {
	out := make([]byte, 0, len(data))

	var inString, escaped, inLine, inBlock bool
	for i := 0; i < len(data); i++ {
		c := data[i]

		switch {
		case inLine:
			if c == '\n' {
				inLine = false
				out = append(out, c)
			}
			continue
		case inBlock:
			if c == '*' && i+1 < len(data) && data[i+1] == '/' {
				inBlock = false
				i++
			}
			continue
		case inString:
			out = append(out, c)
			switch {
			case escaped:
				escaped = false
			case c == '\\':
				escaped = true
			case c == '"':
				inString = false
			}
			continue
		}

		if c == '/' && i+1 < len(data) {
			if data[i+1] == '/' {
				inLine = true
				i++
				continue
			}
			if data[i+1] == '*' {
				inBlock = true
				i++
				continue
			}
		}
		if c == '"' {
			inString = true
		}
		out = append(out, c)
	}

	return stripTrailingCommas(out)
}

// stripTrailingCommas drops a comma followed only by whitespace and a closing
// brace or bracket.
func stripTrailingCommas(data []byte) []byte {
	out := make([]byte, 0, len(data))

	var inString, escaped bool
	for i := 0; i < len(data); i++ {
		c := data[i]

		if inString {
			out = append(out, c)
			switch {
			case escaped:
				escaped = false
			case c == '\\':
				escaped = true
			case c == '"':
				inString = false
			}
			continue
		}
		if c == '"' {
			inString = true
			out = append(out, c)
			continue
		}
		if c == ',' {
			j := i + 1
			for j < len(data) && (data[j] == ' ' || data[j] == '\t' || data[j] == '\n' || data[j] == '\r') {
				j++
			}
			if j < len(data) && (data[j] == '}' || data[j] == ']') {
				continue
			}
		}
		out = append(out, c)
	}

	return out
}
