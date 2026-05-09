# Language-Specific AST-Grep Patterns

Optimized patterns for specific programming languages.

## Python

### Type Hints

```yaml
id: missing-return-type
language: python
rule:
  pattern: 'def $FUNC($$$ARGS):'
  not:
    pattern: 'def $FUNC($$$ARGS) -> $TYPE:'
message: 'Function missing return type annotation'

---
id: missing-param-types
language: python
rule:
  pattern: 'def $FUNC($PARAM):'
  constraints:
    PARAM:
      not:
        regex: '.*:.*'  # No type annotation
message: 'Parameter missing type annotation'
```

### Async Patterns

```yaml
id: sync-in-async
language: python
rule:
  pattern: 'time.sleep($DURATION)'
  inside:
    pattern: 'async def $FUNC($$$ARGS): $$$BODY'
fix: 'await asyncio.sleep($DURATION)'
message: 'Use asyncio.sleep in async functions'

---
id: blocking-io-in-async
language: python
rule:
  any:
    - pattern: 'open($PATH)'
    - pattern: 'requests.get($URL)'
  inside:
    pattern: 'async def $FUNC($$$ARGS): $$$BODY'
message: 'Blocking I/O in async function. Use aiofiles or aiohttp.'
```

### Exception Handling

```yaml
id: bare-except
language: python
severity: warning
rule:
  pattern: |
    except:
        $$$BODY
fix: |
  except Exception as e:
      $$$BODY
message: 'Avoid bare except clauses'

---
id: exception-pass
language: python
rule:
  pattern: |
    except $EXC:
        pass
message: 'Silent exception handling - consider logging'
```

## TypeScript

### Type Safety

```yaml
id: any-type-usage
language: typescript
severity: warning
rule:
  any:
    - pattern: 'const $NAME: any = $VALUE'
    - pattern: 'let $NAME: any = $VALUE'
    - pattern: 'function $FUNC($$$ARGS): any'
message: 'Avoid using "any" type - be more specific'

---
id: type-assertion-warning
language: typescript
rule:
  pattern: '$EXPR as any'
message: 'Type assertion to "any" bypasses type checking'
```

### React Best Practices

```yaml
id: missing-key-prop
language: typescriptreact
rule:
  pattern: '$ARR.map($ITEM => <$COMPONENT $$$PROPS />)'
  not:
    has:
      pattern: 'key={$KEY}'
message: 'Missing key prop in list rendering'

---
id: inline-function-in-jsx
language: typescriptreact
rule:
  pattern: '<$COMPONENT onClick={() => $$$BODY} />'
message: 'Avoid inline functions in JSX - may cause unnecessary re-renders'

---
id: direct-state-mutation
language: typescriptreact
rule:
  pattern: '$STATE.$PROP = $VALUE'
  inside:
    has:
      pattern: 'useState($INIT)'
message: 'Do not mutate state directly - use setState'
```

### Next.js Patterns

```yaml
id: use-next-image
language: typescriptreact
rule:
  pattern: '<img $$$ATTRS />'
fix: '<Image $$$ATTRS />'
message: 'Use next/image for optimized images'

---
id: use-next-link
language: typescriptreact
rule:
  pattern: '<a href="$URL">$$$CHILDREN</a>'
  constraints:
    URL:
      not:
        regex: '^https?://.*'  # External links are OK
fix: '<Link href="$URL">$$$CHILDREN</Link>'
message: 'Use next/link for internal navigation'
```

## Go

### Error Handling

```yaml
id: unchecked-error
language: go
severity: error
rule:
  pattern: '$RESULT, _ := $FUNC($$$ARGS)'
message: 'Error ignored - handle or explicitly ignore with comment'

---
id: error-string-comparison
language: go
rule:
  pattern: 'err.Error() == "$MSG"'
message: 'Use errors.Is() or errors.As() for error comparison'
```

### Concurrency

```yaml
id: goroutine-leak
language: go
rule:
  pattern: 'go $FUNC($$$ARGS)'
  not:
    inside:
      has:
        any:
          - pattern: 'ctx.Done()'
          - pattern: 'select { $$$CASES }'
message: 'Goroutine may leak - ensure proper cancellation'

---
id: mutex-not-deferred
language: go
rule:
  pattern: '$MU.Lock()'
  not:
    follows:
      pattern: 'defer $MU.Unlock()'
message: 'Consider using defer for Unlock()'
```

### Context Usage

```yaml
id: context-first-param
language: go
rule:
  pattern: 'func $FUNC($$$BEFORE, ctx context.Context, $$$AFTER)'
message: 'Context should be the first parameter'
fix: 'func $FUNC(ctx context.Context, $$$BEFORE, $$$AFTER)'
```

## Rust

### Memory Safety

```yaml
id: unsafe-block
language: rust
severity: warning
rule:
  pattern: 'unsafe { $$$BODY }'
message: 'Unsafe block - ensure memory safety is manually verified'

---
id: unwrap-usage
language: rust
rule:
  any:
    - pattern: '$RESULT.unwrap()'
    - pattern: '$OPTION.unwrap()'
message: 'Consider using ? operator or expect() with message'
```

### Ownership Patterns

```yaml
id: clone-in-loop
language: rust
rule:
  pattern: '$VAR.clone()'
  inside:
    any:
      - pattern: 'for $ITEM in $ITER { $$$BODY }'
      - pattern: 'while $COND { $$$BODY }'
message: 'Clone in loop may be inefficient - consider borrowing'
```

## Java

### Resource Management

```yaml
id: resource-not-closed
language: java
rule:
  pattern: '$TYPE $VAR = new $RESOURCE($$$ARGS)'
  not:
    any:
      - inside:
          pattern: 'try ($TYPE $VAR = new $RESOURCE($$$ARGS)) { $$$BODY }'
      - follows:
          pattern: '$VAR.close()'
message: 'Resource may not be closed - use try-with-resources'
```

### Null Safety

```yaml
id: null-check-pattern
language: java
rule:
  pattern: 'if ($OBJ != null) { $$$BODY }'
message: 'Consider using Optional for null handling'

---
id: string-equals
language: java
rule:
  pattern: '$STR.equals($OTHER)'
  not:
    precedes:
      pattern: 'if ($STR != null)'
message: 'Potential NPE - use Objects.equals() or check for null'
fix: 'Objects.equals($STR, $OTHER)'
```

## Language Detection

AST-Grep auto-detects language from file extension:

```bash
# Explicit language specification
sg run --pattern '$PATTERN' --lang python

# Auto-detection (uses file extension)
sg run --pattern '$PATTERN' src/main.py

# Multiple languages
sg run --pattern '$PATTERN' --lang python --lang javascript
```

Supported language identifiers:
- `python`, `javascript`, `typescript`, `typescriptreact`
- `go`, `rust`, `java`, `kotlin`, `scala`
- `c`, `cpp`, `csharp`, `swift`
- `ruby`, `php`, `elixir`, `lua`
- `html`, `css`, `json`, `yaml`
