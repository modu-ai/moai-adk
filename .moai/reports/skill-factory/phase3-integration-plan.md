# PHASE 3: CONTEXT7 INTEGRATION PLAN

## Objective
Enhance skills with Context7 MCP integration for real-time documentation access and AI-powered content delivery.

## Phase 3 Components

### 1. Library Mapping Database

Create comprehensive mapping of libraries to Context7 IDs for automatic documentation fetching.

#### Python Ecosystem
```yaml
python_libraries:
  # Web Frameworks
  fastapi: "/tiangolo/fastapi"
  django: "/django/django"
  flask: "/pallets/flask"
  
  # Data Validation
  pydantic: "/pydantic/pydantic"
  
  # Database
  sqlalchemy: "/sqlalchemy/sqlalchemy"
  tortoise-orm: "/tortoise/tortoise-orm"
  
  # Testing
  pytest: "/pytest-dev/pytest"
  pytest-asyncio: "/pytest-dev/pytest-asyncio"
  
  # Async
  asyncio: "/python/cpython"  # Standard library
  aiohttp: "/aio-libs/aiohttp"
  
  # Authentication
  python-jose: "/mpdavis/python-jose"
  passlib: "/pyca/passlib"
  
  # Utilities
  python-dotenv: "/theskumar/python-dotenv"
  click: "/pallets/click"
```

#### JavaScript/TypeScript Ecosystem
```yaml
javascript_libraries:
  # Frontend Frameworks
  react: "/facebook/react"
  nextjs: "/vercel/next.js"
  vue: "/vuejs/core"
  svelte: "/sveltejs/svelte"
  
  # Build Tools
  vite: "/vitejs/vite"
  webpack: "/webpack/webpack"
  
  # TypeScript
  typescript: "/microsoft/TypeScript"
  
  # Testing
  vitest: "/vitest-dev/vitest"
  jest: "/jestjs/jest"
  playwright: "/microsoft/playwright"
  
  # UI Libraries
  shadcn-ui: "/shadcn/ui"
  radix-ui: "/radix-ui/primitives"
  
  # State Management
  zustand: "/pmndrs/zustand"
  jotai: "/pmndrs/jotai"
```

#### Go Ecosystem
```yaml
go_libraries:
  # Web Frameworks
  gin: "/gin-gonic/gin"
  echo: "/labstack/echo"
  fiber: "/gofiber/fiber"
  
  # Database
  gorm: "/go-gorm/gorm"
  sqlx: "/jmoiron/sqlx"
  
  # Testing
  testify: "/stretchr/testify"
```

#### Rust Ecosystem
```yaml
rust_crates:
  # Web Frameworks
  axum: "/tokio-rs/axum"
  actix-web: "/actix/actix-web"
  rocket: "/SergioBenitez/Rocket"
  
  # Async Runtime
  tokio: "/tokio-rs/tokio"
  
  # Database
  sqlx: "/launchbadge/sqlx"
  diesel: "/diesel-rs/diesel"
```

### 2. Skills Requiring Context7 Integration

#### High Priority (50+ skills)
**Language Skills**:
- moai-lang-python
- moai-lang-typescript
- moai-lang-go
- moai-lang-rust
- moai-lang-javascript

**Domain Skills**:
- moai-domain-backend
- moai-domain-frontend
- moai-domain-database
- moai-domain-testing
- moai-domain-security

**Framework Skills**:
- moai-nextjs-architecture
- moai-react-patterns
- moai-fastapi-advanced
- moai-django-patterns

**BaaS Skills**:
- moai-baas-vercel-ext
- moai-baas-firebase-ext
- moai-baas-supabase-ext

#### Medium Priority (30+ skills)
**Security Skills**:
- moai-security-auth
- moai-security-api
- moai-security-encryption

**Library Skills**:
- moai-lib-shadcn-ui
- moai-lib-radix-ui
- moai-lib-zustand

**Essentials Skills**:
- moai-essentials-debug
- moai-essentials-perf
- moai-essentials-test

### 3. Documentation Section Template

Every skill with Context7 integration should include:

```markdown
## Documentation

For up-to-date library documentation:
Skill("moai-context7-lang-integration")

### Context7 Integration

**Supported Libraries**:
- **[Library Name]**: `/org/project` or `/org/project/version`
  - Topic focus: `api`, `patterns`, `best-practices`
  - Token budget: 3000-5000
  - Usage: Latest stable version documentation

**Example Usage**:
\```python
from moai_adk.integrations import Context7Helper

# Fetch latest documentation
docs = await Context7Helper.get_docs(
    library_id="/tiangolo/fastapi",
    topic="routing patterns",
    tokens=3000
)
\```

**Progressive Token Disclosure**:
- Level 1 (1000 tokens): Core API only
- Level 2 (3000 tokens): API + examples
- Level 3 (5000 tokens): Full documentation
- Level 4 (10000 tokens): Comprehensive reference
```

### 4. Integration Patterns

#### Pattern 1: Real-Time Documentation Fetching
```python
async def get_latest_fastapi_patterns():
    """Fetch latest FastAPI routing patterns from Context7."""
    docs = await get_library_docs(
        context7_library_id="/tiangolo/fastapi",
        topic="routing dependency-injection",
        tokens=3000
    )
    return docs
```

#### Pattern 2: Version-Specific Guidance
```python
async def get_versioned_docs(version: str):
    """Get documentation for specific library version."""
    library_id = f"/tiangolo/fastapi/{version}"
    docs = await get_library_docs(
        context7_library_id=library_id,
        topic="migration-guide",
        tokens=2000
    )
    return docs
```

#### Pattern 3: Multi-Library Integration
```python
async def get_tech_stack_docs():
    """Fetch documentation for entire tech stack."""
    stack = {
        "backend": "/tiangolo/fastapi",
        "database": "/sqlalchemy/sqlalchemy",
        "validation": "/pydantic/pydantic"
    }
    
    docs = {}
    for component, lib_id in stack.items():
        docs[component] = await get_library_docs(
            context7_library_id=lib_id,
            tokens=2000
        )
    return docs
```

### 5. Quality Validation

#### Context7 Integration Checklist
- [ ] Library mapping verified
- [ ] Documentation section added
- [ ] Example usage provided
- [ ] Progressive token disclosure documented
- [ ] Fallback strategy defined
- [ ] Error handling implemented
- [ ] Cache strategy configured
- [ ] Both directories synchronized

### 6. Implementation Steps

**Step 1: Create Library Mapping File**
Location: `.moai/config/context7-libraries.yaml`

**Step 2: Update Skills with Documentation Sections**
- Add Context7 integration section
- Include example usage
- Document supported libraries
- Define token budgets

**Step 3: Test Integration Patterns**
- Verify library resolution
- Test documentation fetching
- Validate caching
- Check error handling

**Step 4: Synchronize Directories**
- Main: `.claude/skills/*/SKILL.md`
- Templates: `src/moai_adk/templates/.claude/skills/*/SKILL.md`

**Step 5: Generate Documentation**
- Update skill catalog
- Create integration guide
- Document best practices

### 7. Success Metrics

**Coverage Targets**:
- 50+ skills with Context7 integration (38% of 132)
- 100% library mapping accuracy
- 95%+ documentation fetch success rate
- <3s average documentation fetch time

**Quality Targets**:
- 100% skills with documentation sections
- 100% directory synchronization
- 0 broken library references
- 100% working examples

### 8. Timeline

**Week 1**: Library mapping + documentation sections (20 skills/day)
**Week 2**: Integration testing + quality validation
**Week 3**: Synchronization + documentation generation
**Week 4**: Final validation + deployment

**Total Duration**: 4 weeks for full Phase 3 completion

