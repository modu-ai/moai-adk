# AST-Grep Pattern Syntax Reference

Complete guide to AST-Grep pattern matching syntax.

## Meta-Variables

### Single Node Capture ($NAME)

Captures exactly one AST node.

```yaml
# Match any function call with one argument
pattern: '$FUNC($ARG)'

# Examples matched:
# - print("hello")
# - len(items)
# - calculate(100)
```

### Variadic Capture ($$$NAME)

Captures zero or more AST nodes.

```yaml
# Match function with any number of arguments
pattern: 'function $NAME($$$ARGS) { $$$BODY }'

# Examples matched:
# - function foo() { return 1; }
# - function bar(a, b, c) { console.log(a); return b + c; }
```

### Anonymous Capture ($$_)

Matches any single node without capturing.

```yaml
# Match if statement regardless of condition
pattern: 'if ($$_) { return $VALUE }'

# Useful when you don't need the matched value
```

### Underscore Wildcard ($_)

Shorthand for anonymous single capture.

```yaml
pattern: '$_.$METHOD($$$ARGS)'
# Matches any method call on any object
```

## Relational Rules

### inside

Match pattern only within another pattern.

```yaml
id: useState-in-component
rule:
  pattern: 'useState($INIT)'
  inside:
    pattern: 'function $COMPONENT($$$PROPS) { $$$BODY }'
    stopBy: end  # Don't search nested functions
```

### has

Pattern must contain another pattern.

```yaml
id: class-with-constructor
rule:
  pattern: 'class $NAME { $$$BODY }'
  has:
    pattern: 'constructor($$$ARGS) { $$$IMPL }'
```

### follows

Pattern must be followed by another pattern.

```yaml
id: error-handling-required
rule:
  pattern: '$ERR := $CALL($$$ARGS)'
  follows:
    pattern: 'if $ERR != nil { $$$BODY }'
```

### precedes

Pattern must be preceded by another pattern.

```yaml
id: declaration-before-use
rule:
  pattern: '$VAR'
  precedes:
    pattern: 'const $VAR = $VALUE'
```

## Composite Rules

### all

All conditions must match.

```yaml
rule:
  all:
    - pattern: 'fetch($URL)'
    - inside:
        pattern: 'async function $NAME() { $$$BODY }'
    - not:
        has:
          pattern: 'try { $$$TRY } catch { $$$CATCH }'
```

### any

At least one condition must match.

```yaml
rule:
  any:
    - pattern: 'console.log($$$ARGS)'
    - pattern: 'console.warn($$$ARGS)'
    - pattern: 'console.error($$$ARGS)'
```

### not

Negates a condition.

```yaml
rule:
  pattern: 'async function $NAME() { $$$BODY }'
  not:
    has:
      pattern: 'await $EXPR'
```

## Stop-By Modifiers

Control how deeply rules search.

```yaml
rule:
  pattern: '$VAR'
  inside:
    pattern: 'function $NAME() { $$$BODY }'
    stopBy:
      rule:
        pattern: 'function $NESTED() { $$$INNER }'
```

Options:
- `end` - Stop at the end of matched node
- `neighbor` - Stop at immediate children
- `rule` - Stop when a specific rule matches

## Regex in Patterns

Use regex for identifier matching.

```yaml
rule:
  pattern: '$FUNC($$$ARGS)'
  constraints:
    FUNC:
      regex: '^(get|fetch|load).*'
```

## Fix Transformations

### Simple Fix

```yaml
fix: 'newFunction($ARGS)'
```

### Multi-line Fix

```yaml
fix: |
  try {
    $ORIGINAL
  } catch (error) {
    console.error(error);
  }
```

### Conditional Fix (via separate rules)

Create separate rules with different severity levels for different fixes.

## Examples

### Detect Deprecated API

```yaml
id: deprecated-substr
language: javascript
rule:
  pattern: '$STR.substr($$$ARGS)'
fix: '$STR.slice($$$ARGS)'
message: 'substr is deprecated, use slice instead'
```

### Enforce Error Handling

```yaml
id: unhandled-promise
language: typescript
rule:
  pattern: '$PROMISE.then($CALLBACK)'
  not:
    has:
      pattern: '.catch($HANDLER)'
message: 'Promise should have error handling'
```

### Security Pattern

```yaml
id: no-eval
language: javascript
severity: error
rule:
  any:
    - pattern: 'eval($CODE)'
    - pattern: 'new Function($$$ARGS)'
message: 'Avoid eval() and new Function() - potential code injection'
```
