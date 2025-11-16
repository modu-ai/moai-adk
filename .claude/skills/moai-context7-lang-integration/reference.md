# Context7 Integration Reference

## API Reference

### mcp__context7__resolve-library-id

**Purpose**: Convert user-friendly library names to Context7-compatible IDs in real-time

**Signature**:
```python
def resolve_library_id(library_name: str) -> str
```

**Parameters**:
- `library_name` (str, required): User-friendly library name
  - Case-insensitive: "fastapi", "FastAPI", "FASTAPI" all valid
  - Examples: "fastapi", "django", "react", "gin", "sqlalchemy"
  - Length: 3-50 characters recommended

**Returns**:
- `str`: Context7-compatible library ID
  - Format: `/org/project` or `/org/project/version`
  - Examples:
    - `/tiangolo/fastapi`
    - `/tiangolo/fastapi/v0.115.0`
    - `/django/django`
    - `/facebook/react`

**Raises**:
- `LibraryNotFoundError`: Library not recognized by Context7
  - Solution: Check spelling or use manual documentation link
- `ResolverTimeoutError`: Query timeout (exceeds 5 seconds)
  - Solution: Retry with exponential backoff
- `InvalidLibraryNameError`: Name format invalid (empty, too long)
  - Solution: Validate input before calling

**Examples**:
```python
# Basic usage
lib_id = resolve_library_id("fastapi")
# Returns: /tiangolo/fastapi or /tiangolo/fastapi/v0.115.0

# Case-insensitive
lib_id = resolve_library_id("FastAPI")
# Returns: /tiangolo/fastapi

# Error handling
try:
    lib_id = resolve_library_id("nonexistent-lib")
except LibraryNotFoundError:
    print("Library not found in Context7")
    lib_id = None  # Use fallback documentation

# Alternative names
lib_id = resolve_library_id("django")  # /django/django
lib_id = resolve_library_id("Django")  # /django/django (case insensitive)
```

**Performance Notes**:
- First call may take 1-2 seconds (API query)
- Subsequent calls within 30 days use cache (instant)
- Cache TTL: 30 days for stable versions
- Cache TTL: 7 days for beta/latest versions

---

### mcp__context7__get-library-docs

**Purpose**: Fetch documentation for a specific library with token optimization

**Signature**:
```python
def get_library_docs(
    context7_compatible_library_id: str,
    tokens: int = 5000,
    topic: str = None
) -> str
```

**Parameters**:
- `context7_compatible_library_id` (str, required): Output from resolve_library_id
  - Format: `/org/project` or `/org/project/version`
  - Example: `/tiangolo/fastapi/v0.115.0`

- `tokens` (int, optional): Maximum tokens to fetch
  - Default: 5000 (balanced)
  - Range: 1000-10000
  - Recommended values:
    - 1000: Quick reference/summary only
    - 3000: API docs + examples
    - 5000: Full documentation (default)
    - 10000: Comprehensive + advanced topics
  - Higher values = slower fetch + higher token cost

- `topic` (str, optional): Focus on specific documentation topic
  - Examples: "routing", "async", "dependency-injection", "middleware"
  - If None: Returns overview/table of contents
  - Specific topic + lower tokens = faster, cheaper requests

**Returns**:
- `str`: Markdown-formatted documentation
  - Includes version information
  - Includes relevant links and examples
  - UTF-8 encoded
  - Line-break format: Unix style (LF)

**Raises**:
- `DocumentationNotFoundError`: Docs unavailable for library/version
  - Solution: Try with different topic or lower token count
- `TokenLimitExceededError`: tokens > 10000
  - Solution: Reduce token count, split into multiple requests
- `InvalidLibraryIDError`: Malformed library ID
  - Solution: Verify output from resolve_library_id
- `NetworkTimeoutError`: API response timeout (>30 seconds)
  - Solution: Retry with exponential backoff, use cached version

**Examples**:
```python
# Basic usage: Get full documentation
docs = get_library_docs("/tiangolo/fastapi")
print(docs)

# With topic focus: Get only routing docs
docs = get_library_docs(
    context7_compatible_library_id="/tiangolo/fastapi",
    topic="routing",
    tokens=3000
)
print(docs)

# Minimal: Quick reference only
docs = get_library_docs(
    context7_compatible_library_id="/tiangolo/fastapi",
    tokens=1000
)
print(docs[:500])

# With error handling
try:
    docs = get_library_docs(
        context7_compatible_library_id="/tiangolo/fastapi",
        topic="dependency-injection",
        tokens=4000
    )
except DocumentationNotFoundError:
    # Try broader topic
    docs = get_library_docs(
        context7_compatible_library_id="/tiangolo/fastapi",
        topic="api",
        tokens=3000
    )
```

**Performance Notes**:
- Fetch time: 500ms - 3s depending on token count
- 1000 tokens: ~500ms, ~2KB
- 5000 tokens: ~1.5s, ~10KB
- 10000 tokens: ~3s, ~20KB
- Cache duration: 30 days (for stable versions)

---

## Available Libraries

### Python Ecosystem

**Web Frameworks**:
- FastAPI: `/tiangolo/fastapi` - Modern async web framework
- Django: `/django/django` - Full-featured web framework
- Flask: `/pallets/flask` - Lightweight web framework

**Data & ORM**:
- SQLAlchemy: `/sqlalchemy/sqlalchemy` - SQL toolkit and ORM
- Pydantic: `/pydantic/pydantic` - Data validation library
- Alembic: `/sqlalchemy/alembic` - Database migrations

**Testing & Quality**:
- pytest: `/pytest-dev/pytest` - Testing framework
- Ruff: `/astral-sh/ruff` - Fast Python linter

**Async & Concurrency**:
- asyncio: `/python/asyncio` - Built-in async library
- httpx: `/encode/httpx` - Async HTTP client

### JavaScript/TypeScript Ecosystem

**Frameworks & Meta**:
- Next.js: `/vercel/next.js` - React framework with SSR
- React: `/facebook/react` - UI library
- Vue: `/vuejs/vue` - Progressive framework

**Type Systems**:
- TypeScript: `/microsoft/TypeScript` - Typed JavaScript
- Zod: `/colinhacks/zod` - TypeScript-first validation

**Testing & Build**:
- Vitest: `/vitest-dev/vitest` - Unit test framework
- Vite: `/vitejs/vite` - Build tool
- Jest: `/jestjs/jest` - Testing framework

**HTTP & APIs**:
- Axios: `/axios/axios` - HTTP client
- tRPC: `/trpc/trpc` - End-to-end typesafe APIs

### Go Ecosystem

**Web Frameworks**:
- Gin: `/gin-gonic/gin` - Fast web framework
- Echo: `/labstack/echo` - Minimalist framework
- GORM: `/go-gorm/gorm` - ORM library

**CLI & Tools**:
- Cobra: `/spf13/cobra` - CLI framework
- Viper: `/spf13/viper` - Configuration library

**Database**:
- sqlc: `/sqlc-dev/sqlc` - SQL compiler
- PgX: `/jackc/pgx` - PostgreSQL driver

**HTTP & RPC**:
- rpc: `/tmc/grpc` - gRPC framework
- Gorilla: `/gorilla/mux` - URL router

### Other Languages

**Swift**:
- Swift: `/apple/swift` - Apple's programming language

**Kotlin**:
- Kotlin: `/jetbrains/kotlin` - JVM-based language

**Rust**:
- Rust: `/rust-lang/rust` - Systems programming language

**Ruby**:
- Rails: `/rails/rails` - Web framework
- RSpec: `/rspec/rspec` - Testing framework

---

## Best Practices

### 1. Cache Aggressively

**Strategy**:
- Always cache successful fetches
- TTL: 30 days for stable versions
- TTL: 7 days for beta/latest
- Clear cache only on version mismatch

**Implementation**:
```python
# Check cache before fetching
cached_docs = cache.get(library_id, topic)
if cached_docs:
    return cached_docs  # Instant return

# Fetch and cache
docs = get_library_docs(library_id, topic)
cache.set(library_id, docs, topic)
return docs
```

### 2. Handle Errors Gracefully

**Strategy**:
- Always provide fallback documentation link
- Log all failures for debugging
- Retry with backoff (3 retries maximum)
- Fail open (show manual link) rather than crash

**Implementation**:
```python
for attempt in range(3):
    try:
        return get_library_docs(library_id, topic)
    except DocumentationNotFoundError:
        if attempt < 2:
            topic = "api"  # Broaden topic on retry
        else:
            return f"See: https://docs.example.com/{library_id}"
```

### 3. Optimize Token Usage

**Strategy**:
- Request specific topic, not full documentation
- Use minimum viable tokens (1000-3000)
- Cache liberally to avoid re-fetches
- Monitor token consumption

**Best Practices**:
- Specific topic + 1000 tokens: Quick reference
- General API + 3000 tokens: Balanced documentation
- Full docs + 5000 tokens: Comprehensive reference
- Advanced topics + 10000 tokens: Full reference

### 4. Version Management

**Strategy**:
- Pin to specific versions in production
- Use latest for documentation-only context
- Maintain compatibility matrix
- Test across versions

**Implementation**:
```python
# Production: Pin version
library_id = "/tiangolo/fastapi/v0.115.0"

# Documentation: Use latest
library_id = "/tiangolo/fastapi"  # Latest version

# Compatibility matrix
versions = ["v0.100.0", "v0.110.0", "v0.115.0"]
for version in versions:
    lib_id = f"/tiangolo/fastapi/{version}"
    docs = get_library_docs(lib_id, topic="breaking-changes")
```

### 5. Integration Patterns

**Strategy**:
- Load Skill in all Language Skills
- Call via `Skill("moai-context7-lang-integration")`
- Document library references in each Skill
- Keep integration code minimal and reusable

**Pattern in Language Skills**:
```markdown
## External Documentation

For up-to-date library documentation:
Skill("moai-context7-lang-integration")

```python
library_id = resolve_library_id("fastapi")
docs = get_library_docs(library_id, topic="routing")
```
```

---

## Troubleshooting

### Problem: "LibraryNotFoundError"

**Symptoms**: Skill returns error that library not recognized

**Common Causes**:
1. Typo in library name ("fastapi" vs "fastpai")
2. Case sensitivity (shouldn't matter, but verify)
3. Library not in Context7 registry
4. Using unofficial fork/fork name

**Solution Steps**:
1. Verify spelling against official Context7 library list
2. Try alternative names:
   - "fastapi" → "FastAPI" → "fast-api"
   - "django" → "Django" → "django-web"
3. Use manual documentation link as fallback
4. Check Context7 status page for outages

**Workaround**:
```python
def safe_resolve(library_name: str) -> Optional[str]:
    alternatives = {
        "fastapi": ["FastAPI", "fast-api"],
        "django": ["Django"]
    }

    names_to_try = [library_name] + alternatives.get(library_name, [])

    for name in names_to_try:
        try:
            return resolve_library_id(name)
        except LibraryNotFoundError:
            continue

    return None  # Fallback to manual link
```

### Problem: "DocumentationNotFoundError"

**Symptoms**: Documentation unavailable for library/version

**Common Causes**:
1. Documentation not available for specified version
2. Library is too new (docs not indexed yet)
3. Topic doesn't exist for this library
4. Version has been deprecated

**Solution Steps**:
1. Try with different topic (specific → general)
2. Reduce token count and retry
3. Check if library version is supported
4. Use previous cached version

**Workaround**:
```python
def get_docs_with_fallback(library_id: str, topic: str) -> str:
    topics_to_try = [topic, "api", "overview", None]

    for current_topic in topics_to_try:
        try:
            return get_library_docs(library_id, topic=current_topic, tokens=3000)
        except DocumentationNotFoundError:
            continue

    return "Documentation not available"
```

### Problem: Performance Issues

**Symptoms**: Slow fetch times, timeouts, user experience degradation

**Common Causes**:
1. No caching (fetching same docs repeatedly)
2. Too many tokens requested
3. Fetching entire docs when only need snippet
4. Too many parallel requests (rate limiting)
5. Context7 API service degradation

**Solution Steps**:
1. Enable caching with 30-day TTL
2. Reduce token count (1000-3000 sufficient)
3. Use specific topic instead of full docs
4. Implement batch requests (not parallel)
5. Check Context7 API status

**Optimization Checklist**:
- [ ] Caching enabled for 30+ days
- [ ] Token count: 1000-3000 (not 10000)
- [ ] Using specific topic parameter
- [ ] Sequential, not parallel requests
- [ ] Monitoring Context7 API status

### Problem: Token Budget Exceeded

**Symptoms**: TokenBudgetExceededError or allocation failures

**Common Causes**:
1. Too many concurrent documentation fetches
2. Token allocation percentages > 100%
3. Repeatedly fetching without caching
4. Requesting too many tokens per fetch

**Solution Steps**:
1. Implement token budget manager
2. Allocate reasonable percentages (40%+35%+25% = 100%)
3. Use caching aggressively
4. Reduce tokens per request
5. Batch requests instead of parallel

**Budget Formula**:
```
Total Budget: 20,000 tokens/session
Skill 1: 40% = 8,000 tokens
Skill 2: 35% = 7,000 tokens
Skill 3: 25% = 5,000 tokens

Per fetch: Limit to 3,000 tokens
Cached hits: 0 tokens (no API call)
Budget: 20,000 tokens / 3,000 per fetch = ~6 fetches max
With caching: Effectively unlimited (cache hits = 0 tokens)
```

### Problem: Version Mismatch

**Symptoms**: Incorrect or outdated documentation returned

**Common Causes**:
1. Not specifying version (gets latest)
2. Version string format incorrect
3. Requesting latest when need specific version
4. Cache stale after version upgrade

**Solution Steps**:
1. Pin to specific version when needed
2. Use correct version format (v0.115.0, not 0.115.0)
3. Clear cache after version upgrade
4. Verify version string format

**Correct Version Formats**:
```python
# Correct
lib_id = "/tiangolo/fastapi/v0.115.0"    # Good
lib_id = "/tiangolo/fastapi"              # Latest
lib_id = "/tiangolo/fastapi/v0.100.0"    # Specific

# Incorrect (these will fail)
lib_id = "/tiangolo/fastapi/0.115.0"      # Missing 'v'
lib_id = "/tiangolo/fastapi@0.115.0"      # Wrong separator
lib_id = "tiangolo/fastapi/v0.115.0"      # Missing leading /
```

---

## Related Skills & Resources

**Complementary Skills**:
- `Skill("moai-lang-python")` - Python-specific documentation
- `Skill("moai-lang-typescript")` - TypeScript-specific documentation
- `Skill("moai-lang-go")` - Go-specific documentation
- `Skill("moai-domain-backend")` - Backend API patterns
- `Skill("moai-domain-frontend")` - Frontend framework patterns

**External Resources**:
- [Context7 Official Documentation](https://context7.io)
- [Context7 Library Registry](https://context7.io/libraries)
- [MCP Protocol Specification](https://modelcontextprotocol.io)
- [MCP Documentation](https://modelcontextprotocol.io/docs)

**Configuration**:
- Cache location: `.moai/cache/context7/`
- Cache index: `.moai/cache/context7/index.json`
- Token budget: `.moai/config/config.json` → `token_budget`

