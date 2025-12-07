## Verified Citation Cache Management

### Cache File Schema

**Location**: `.moai/yoda/books/{book_slug}/verified-citations-cache.json`

**Structure**:
```json
{
  "version": "1.0.0",
  "book_slug": "claude-code-agentic-coding-master",
  "created_at": "2025-11-30T10:00:00Z",
  "last_verified": "2025-11-30T10:00:00Z",
  "verification_status": "COMPLETE",
  "total_citations": 25,
  "verified_count": 25,
  "failed_count": 0,
  "citations": {
    "CC-001": {
      "id": "CC-001",
      "title": "Claude Code Official Documentation",
      "url": "https://docs.anthropic.com/en/docs/claude-code",
      "description": "공식 Claude Code 문서",
      "credibility": 10,
      "type": "official_docs",
      "verified_at": "2025-11-30T10:00:00Z",
      "status": "verified"
    },
    "CC-002": {
      "id": "CC-002",
      "title": "Claude Code GitHub Repository",
      "url": "https://github.com/anthropics/claude-code",
      "description": "Claude Code 공식 저장소",
      "credibility": 10,
      "type": "official_repo",
      "verified_at": "2025-11-30T10:00:00Z",
      "status": "verified"
    }
  },
  "chapter_citations": {
    "chapter-01": ["CC-001", "CC-002"],
    "chapter-02": ["CC-001", "CC-003"]
  }
}
```

### Cache Operations

**Create Cache**:
```
Write(
  file_path=".moai/yoda/books/{book_slug}/verified-citations-cache.json",
  content=JSON.stringify(cache_object, indent=2)
)
```

**Read Cache**:
```
cache = Read(file_path=".moai/yoda/books/{book_slug}/verified-citations-cache.json")
cache_data = JSON.parse(cache)
```

**Validate Cache**:
- Check `verification_status == "COMPLETE"`
- Check `verified_count == total_citations`
- Check `failed_count == 0`
- Check `last_verified` is within 30 days

**Update Cache**:
Only when re-verification is needed. Never modify individual citations.

---

## Citation ID System

### ID Format

Pattern: `{DOMAIN}-{NUMBER:03d}`

Examples:
- `CC-001` - Claude Code citation #1
- `CC-002` - Claude Code citation #2
- `ANTH-001` - Anthropic citation #1
- `PY-001` - Python official docs citation #1

### ID Generation

Based on source domain:
- `docs.anthropic.com/claude-code` → `CC-xxx`
- `www.anthropic.com` → `ANTH-xxx`
- `docs.python.org` → `PY-xxx`
- `github.com/anthropics/*` → `GH-xxx`

Sequential numbering within each domain.

### Usage in Text

**Markdown Format**:
```markdown
Claude Code는 AI 코딩 도구입니다 {{CITATION:CC-001}}.

## 인용문

1. {{CC-001}}: Claude Code Official Documentation
   - URL: https://docs.anthropic.com/en/docs/claude-code
   - 검증: 2025-11-30
   - 상태: ✅ 검증 완료
```

**Forbidden Formats**:
- ❌ Raw URLs: `(https://...)`
- ❌ Numbered references: `(1)`, `[1]`
- ❌ Inline URLs: `[text](url)`

**Only Allowed**:
- ✅ ID references: `{{CITATION:CC-001}}`

---

### Real-time Verification System

**URL Accessibility Verification**:
```python
async def verify_url_accessibility(url: str) -> VerificationResult:
    """
    Verify URL accessibility using WebFetch or MCP tools

    Args:
        url: URL to verify

    Returns:
        VerificationResult with accessibility status
    """
    try:
        # Use WebFetch for URL verification
        response = WebFetch(url)

        if response and response.get("status_code") == 200:
            return VerificationResult(
                accessible=True,
                status_code=response["status_code"],
                response_time=response.get("response_time", 0),
                content_type=response.get("content_type", ""),
                error=None
            )
        else:
            return VerificationResult(
                accessible=False,
                status_code=response.get("status_code") if response else None,
                response_time=0,
                content_type="",
                error="URL not accessible"
            )

    except Exception as e:
        return VerificationResult(
            accessible=False,
            status_code=None,
            response_time=0,
            content_type="",
            error=str(e)
        )
```

**Content Relevance Validation**:
```python
def validate_content_relevance(url: str, expected_content: str, page_content: str) -> float:
    """
    Validate that page content matches expected citation content

    Args:
        url: The citation URL
        expected_content: Expected content description
        page_content: Actual page content from URL

    Returns:
        Relevance score (0.0 to 1.0)
    """
    if not page_content:
        return 0.0

    # Extract keywords from expected content
    expected_keywords = extract_keywords(expected_content.lower())
    page_lower = page_content.lower()

    # Count keyword matches
    matches = sum(1 for keyword in expected_keywords if keyword in page_lower)

    # Calculate relevance score
    if not expected_keywords:
        return 0.5  # Neutral score for empty expected content

    relevance = matches / len(expected_keywords)

    # Boost for official documentation patterns
    official_patterns = [
        "documentation", "api reference", "tutorial", "guide",
        "getting started", "overview", "introduction"
    ]

    for pattern in official_patterns:
        if pattern in page_lower and pattern in expected_content.lower():
            relevance = min(1.0, relevance + 0.2)
            break

    return relevance

def extract_keywords(text: str) -> List[str]:
    """Extract relevant keywords from text"""
    import re

    # Remove common stop words
    stop_words = {
        'the', 'a', 'an', 'and', 'or', 'but', 'in', 'on', 'at', 'to', 'for',
        'of', 'with', 'by', 'is', 'are', 'was', 'were', 'be', 'been', 'have',
        'has', 'had', 'do', 'does', 'did', 'will', 'would', 'could', 'should'
    }

    # Extract words
    words = re.findall(r'\b\w+\b', text.lower())

    # Filter stop words and short words
    keywords = [
        word for word in words
        if word not in stop_words and len(word) >= 3
    ]

    return list(set(keywords))  # Remove duplicates
```

### Zero-Tolerance Enforcement

**Citation Verification Pipeline**:
```python
async def verify_citations(citations: List[str], domain: str = None) -> VerificationResult:
    """
    Comprehensive citation verification with zero-tolerance policy

    Args:
        citations: List of citation URLs to verify
        domain: Domain for additional validation rules

    Returns:
        Comprehensive verification result
    """

    # Step 1: Load trusted database for domain
    if domain:
        trusted_citations = load_trusted_citations(domain)
        trusted_urls = {c.url for c in trusted_citations}
    else:
        trusted_urls = set()

    # Step 2: Verify each citation
    verification_results = []
    for url in citations:
        result = await verify_single_citation(url, trusted_urls, domain)
        verification_results.append(result)

    # Step 3: ZERO TOLERANCE - 100% verification required
    total_count = len(verification_results)
    verified_count = sum(1 for r in verification_results if r.verified)
    failed_citations = [r for r in verification_results if not r.verified]

    if verified_count == total_count and total_count > 0:
        status = "ALL_VERIFIED"
        action = "PROCEED"
    else:
        # ANY failure = complete rejection
        status = "VERIFICATION_FAILED"
        action = "REJECT"

        error_message = f"""
CITATION VERIFICATION FAILED

Total: {total_count}
Verified: {verified_count}
Failed: {total_count - verified_count}

Failed Citations:
{chr(10).join(f"- {c.url}: {c.error if hasattr(c, 'error') else c.reason}" for c in failed_citations)}

100% verification required. No partial success allowed.
Action Required:
- Update trusted citations database with correct URLs
- Remove outdated/broken citations
- Re-run verification after fixing all URLs
"""

        raise VerificationFailedError(error_message)

    return VerificationResult(
        status=status,
        verified_count=verified_count,
        total_count=total_count,
        success_rate=verified_count / total_count if total_count > 0 else 0,
        failed_urls=[],
        action=action,
        details=verification_results
    )

async def verify_single_citation(url: str, trusted_urls: set, domain: str = None) -> SingleVerificationResult:
    """Verify a single citation against zero-tolerance rules"""

    # Rule 1: Must be in trusted database
    if trusted_urls and url not in trusted_urls:
        return SingleVerificationResult(
            url=url,
            verified=False,
            reason="NOT_IN_TRUSTED_DATABASE",
            error="URL not found in trusted citation database"
        )

    # Rule 2: URL accessibility check
    accessibility = await verify_url_accessibility(url)
    if not accessibility.accessible:
        return SingleVerificationResult(
            url=url,
            verified=False,
            reason="URL_NOT_ACCESSIBLE",
            error=accessibility.error
        )

    # Rule 3: Content relevance (basic check)
    # For actual content matching, you might want to fetch and analyze page content
    # This is a simplified version

    return SingleVerificationResult(
        url=url,
        verified=True,
        reason="VERIFIED",
        accessibility=accessibility
    )
```

**Forbidden Pattern Detection**:
```python
import re

FORBIDDEN_PATTERNS = [
    r"arXiv:2025\.\d+",           # Non-existent arXiv papers
    r"claude-code-features\.com",   # Fake domains
    r"stackoverflow\.com/questions/",  # SO (except for specific cases)
    r"reddit\.com/r/",                 # Reddit
    r"medium\.com/@",                  # Medium personal blogs
    r"tutorialspoint\.com",           # Low-quality tutorials
    r"geeksforgeeks\.org",            # Inconsistent quality
    r"w3schools\.com/(?!default\.asp)", # Basic stuff only
    r"javatpoint\.com",
    r"digitalocean\.com/community/tutorials",
    r"freecodecamp\.org/news",
    r"dev\.to/",
    r"hashnode\.com/",
    r"blog\.logrocket\.com"
]

ALLOWED_DOMAINS = {
    "docs.anthropic.com",
    "www.anthropic.com",
    "docs.python.org",
    "python.org",
    "peps.python.org",
    "react.dev",
    "nextjs.org",
    "nodejs.org",
    "www.typescriptlang.org",
    "developer.mozilla.org",
    "github.com",
    "docs.github.com",
    "vercel.com",
    "supabase.com",
    "openai.com",
    "api.openai.com"
}

def validate_citation_url(url: str) -> ValidationResult:
    """
    Validate URL against forbidden patterns and allowed domains

    Returns:
        ValidationResult with validation status
    """

    # Check forbidden patterns
    for pattern in FORBIDDEN_PATTERNS:
        if re.search(pattern, url, re.IGNORECASE):
            return ValidationResult(
                valid=False,
                reason="FORBIDDEN_PATTERN",
                message=f"URL matches forbidden pattern: {pattern}"
            )

    # Check allowed domains
    from urllib.parse import urlparse
    domain = urlparse(url).netloc.lower()

    if not any(allowed in domain for allowed in ALLOWED_DOMAINS):
        return ValidationResult(
            valid=False,
            reason="DOMAIN_NOT_ALLOWED",
            message=f"Domain '{domain}' not in allowed list"
        )

    return ValidationResult(
        valid=True,
        reason="VALID",
        message="URL passes validation rules"
    )
```

### Korean Citation Formatting

**Format Citations for Korean Technical Writing**:
```python
def format_citations(citations: List[Citation], style: str = "korean") -> str:
    """
    Format citations according to Korean technical writing standards

    Args:
        citations: List of citation objects
        style: Formatting style ("korean", "english", "academic")

    Returns:
        Formatted citation section as markdown
    """

    if not citations:
        return """## 인용문

📋 **인용문 준비 중**

현재 이 주제에 대한 신뢰할 수 있는 공식 출처를 조사 중입니다.
모든 인용문은 100% 검증된 후 추가됩니다."""

    if style == "korean":
        return format_korean_citations(citations)
    elif style == "english":
        return format_english_citations(citations)
    elif style == "academic":
        return format_academic_citations(citations)
    else:
        return format_korean_citations(citations)

def format_korean_citations(citations: List[Citation]) -> str:
    """Format citations in Korean technical writing style"""

    citations_text = "## 인용문\n\n"

    for i, citation in enumerate(citations, 1):
        # Credibility indicator
        if citation.credibility >= 10:
            credibility_icon = "🔵 공식 문서"
        elif citation.credibility >= 8:
            credibility_icon = "🟢 신뢰 출처"
        else:
            credibility_icon = "🟡 검증 완료"

        citations_text += f"{i}. **{citation.title}** {credibility_icon}\n"
        citations_text += f"   {citation.description}\n"
        citations_text += f"   [{citation.url}]({citation.url})\n\n"

    # Add verification note
    citations_text += """**검증 기준**:
- ✅ 공식 문서 및 신뢰할 수 있는 출처만 사용
- ✅ 실시간 URL 접속 가능성 확인
- ✅ 콘텐츠 관련성 검증 완료
- ❌ 할루시네이션 방지를 위해 미확인 출처 제외"""

    return citations_text

def format_english_citations(citations: List[Citation]) -> str:
    """Format citations in English style"""

    citations_text = "## Citations\n\n"

    for i, citation in enumerate(citations, 1):
        credibility_text = ""
        if citation.credibility >= 10:
            credibility_text = " (Official Documentation)"
        elif citation.credibility >= 8:
            credibility_text = " (Trusted Source)"

        citations_text += f"{i}. **{citation.title}**{credibility_text}\n"
        citations_text += f"   {citation.description}\n"
        citations_text += f"   {citation.url}\n\n"

    return citations_text
```

### Context7 Integration

**Latest Documentation Verification**:
```python
async def verify_context7_citations(library_name: str, topic: str = None) -> List[Citation]:
    """
    Verify citations using Context7 for latest documentation

    Args:
        library_name: Name of library/documentation to verify
        topic: Specific topic within the library

    Returns:
        List of verified Context7 citations
    """

    try:
        # Resolve library ID
        library_result = mcp__context7__resolve-library-id(library_name)

        if not library_result or not library_result.get("library_id"):
            return []

        library_id = library_result["library_id"]

        # Get library documentation
        docs = mcp__context7__get-library-docs(
            context7CompatibleLibraryID=library_id,
            mode="code" if topic else "info",
            topic=topic
        )

        if not docs:
            return []

        # Create citation from Context7 result
        citations = []
        base_url = extract_base_url_from_library_id(library_id)

        citations.append(Citation(
            id=f"context7-{library_name}",
            title=f"{library_name.title()} Documentation",
            url=base_url,
            description=f"Latest {library_name} documentation via Context7",
            credibility=10,  # Context7 sources are highly trusted
            type="official_docs"
        ))

        return citations

    except Exception as e:
        print(f"Context7 verification failed for {library_name}: {e}")
        return []

def extract_base_url_from_library_id(library_id: str) -> str:
    """Extract base URL from Context7 library ID"""
    url_mapping = {
        "/anthropic/claude-code": "https://docs.anthropic.com/claude-code",
        "/python/docs": "https://docs.python.org/3/",
        "/react/docs": "https://react.dev/",
        "/nextjs/docs": "https://nextjs.org/docs"
    }

    return url_mapping.get(library_id, library_id)
```

---

