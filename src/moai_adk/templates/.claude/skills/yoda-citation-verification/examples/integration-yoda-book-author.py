#!/usr/bin/env python3
"""
YodA Book Author Integration Example

This example shows how the yoda-citation-verification skill integrates
seamlessly with the yoda-book-author agent to provide foolproof
citation verification.
"""

import asyncio
from typing import Dict, List, Any
from dataclasses import dataclass

@dataclass
class ChapterGenerationRequest:
    """Request structure for chapter generation with citation verification"""
    topic: str
    chapter_title: str
    chapter_num: int
    target_audience: str
    domain: str
    prev_summary: str = None

class YodaBookAuthorWithCitationVerification:
    """
    Enhanced YodA Book Author with integrated citation verification

    This class demonstrates the seamless integration between yoda-book-author
    and yoda-citation-verification skills for zero-tolerance citation management.
    """

    def __init__(self):
        self.citation_verifier = None  # Would be initialized with CitationVerifier
        self.error_recovery = None     # Would be initialized with CitationErrorRecovery

    async def generate_chapter_with_verified_citations(
        self, request: ChapterGenerationRequest
    ) -> Dict[str, Any]:
        """
        Generate chapter content with foolproof citation verification

        This is the main integration point that shows how the citation verification
        skill works with the yoda-book-author agent.
        """

        print(f"🚀 Starting chapter generation: {request.chapter_title}")
        print(f"📋 Domain: {request.domain} | Topic: {request.topic}")

        # STEP 1: Generate chapter content WITHOUT citations (zero-tolerance rule)
        print("\n📝 Step 1: Generating chapter content (no citations)")
        chapter_content = await self._generate_content_without_citations(request)

        # STEP 2: Load trusted citations for the domain
        print("\n🔍 Step 2: Loading trusted citations")
        trusted_citations = await self._load_trusted_citations(request.domain)

        if not trusted_citations:
            print("⚠️ No trusted citations found - using preparation message")
            citations_section = self._generate_citation_preparation_message()
            verification_status = "NO_SOURCES"
        else:
            print(f"✅ Found {len(trusted_citations)} trusted citations")

            # STEP 3: Verify all citations in real-time
            print("\n🔐 Step 3: Real-time citation verification")
            verification_result = await self._verify_citations_realtime(trusted_citations)

            if verification_result["status"] == "ALL_VERIFIED":
                print("✅ All citations verified successfully")
                citations_section = self._format_verified_citations(trusted_citations)
                verification_status = "VERIFIED"

            elif verification_result["status"] == "PARTIAL_FAILURE":
                print(f"⚠️ Partial verification failure: {verification_result['verified_count']}/{len(trusted_citations)}")

                # Attempt recovery
                recovered_citations = await self._recover_from_verification_failure(
                    verification_result["failed_urls"], request.domain
                )

                if recovered_citations:
                    print(f"🔄 Recovery successful: {len(recovered_citations)} alternative citations")
                    citations_section = self._format_verified_citations(recovered_citations)
                    verification_status = "RECOVERED"
                else:
                    citations_section = self._generate_citation_preparation_message()
                    verification_status = "RECOVERY_FAILED"

            else:  # COMPLETE_FAILURE
                print("❌ Complete citation verification failure")
                citations_section = self._generate_citation_preparation_message()
                verification_status = "FAILED"

        # STEP 4: Combine content and citations
        print("\n📚 Step 4: Combining content with verified citations")
        final_chapter = self._combine_content_and_citations(chapter_content, citations_section)

        # STEP 5: Final validation
        print("\n✅ Step 5: Final validation")
        validation_result = await self._final_validation(final_chapter, trusted_citations)

        return {
            "chapter_content": final_chapter,
            "citations_count": len(trusted_citations) if trusted_citations else 0,
            "verification_status": verification_status,
            "validation_passed": validation_result["passed"],
            "processing_time": validation_result["processing_time"],
            "warnings": validation_result.get("warnings", [])
        }

    async def _generate_content_without_citations(self, request: ChapterGenerationRequest) -> str:
        """Generate chapter content without any citations (zero-tolerance rule)"""

        # This would integrate with the actual content generation system
        # For this example, we'll return a template

        content = f"""# {request.chapter_num}장. {request.chapter_title}

## 학습 목표

이 장에서는 다음을 배웁니다:

- {request.topic}의 기본 개념과 원리
- 실제 개발 환경에서의 활용 방법
- 모범 사례와 주의사항

## 핵심 내용

### 기본 개념

{request.topic}은 현대 개발에서 매우 중요한 역할을 합니다. 공식 문서에 따르면,
이 기술은 개발자 생산성을 크게 향상시킬 수 있습니다.

### 주요 특징

1. **자동화**: 반복적인 작업을 자동으로 처리
2. **통합**: 다양한 도구와 시스템과의 원활한 연동
3. **확장성**: 필요에 따른 기능 확장 가능

### 실제 적용

실제 프로젝트에서 {request.topic}을 적용할 때는 다음 사항을 고려해야 합니다:

- 프로젝트 규모와 복잡성
- 팀의 기술 수준
- 장기 유지보수 계획

## 다음 장에서는

이 장에서 배운 내용을 바탕으로 다음 장에서는 더 고급 기능과 실전 프로젝트에 대해 알아보겠습니다.
"""

        return content

    async def _load_trusted_citations(self, domain: str) -> List[Dict]:
        """Load trusted citations from the database"""

        # This would use the actual load_trusted_citations function
        # For this example, we'll return mock data

        mock_citations = {
            "claude_code": [
                {
                    "id": "claude-code-docs-main",
                    "title": "Claude Code Documentation",
                    "url": "https://docs.anthropic.com/claude-code",
                    "description": "공식 Claude Code 문서 - 설치, 설정, 기능",
                    "credibility": 10,
                    "type": "official_docs"
                },
                {
                    "id": "claude-code-getting-started",
                    "title": "Claude Code Getting Started",
                    "url": "https://docs.anthropic.com/claude-code/getting-started",
                    "description": "Claude Code 시작 가이드",
                    "credibility": 10,
                    "type": "official_tutorial"
                }
            ],
            "python": [
                {
                    "id": "python-docs-main",
                    "title": "Python Documentation",
                    "url": "https://docs.python.org/3/",
                    "description": "Python 3 공식 문서",
                    "credibility": 10,
                    "type": "official_docs"
                }
            ]
        }

        return mock_citations.get(domain, [])

    async def _verify_citations_realtime(self, citations: List[Dict]) -> Dict[str, Any]:
        """Verify citations in real-time using WebFetch or MCP tools"""

        # This would use the actual verify_citations function
        # For this example, we'll simulate verification

        import random

        # Simulate some verification results
        verified_count = sum(1 for _ in citations if random.random() > 0.2)  # 80% success rate

        if verified_count == len(citations):
            return {"status": "ALL_VERIFIED", "verified_count": verified_count}
        elif verified_count > 0:
            failed_urls = [c["url"] for c in citations[verified_count:]]
            return {
                "status": "PARTIAL_FAILURE",
                "verified_count": verified_count,
                "failed_urls": failed_urls
            }
        else:
            return {
                "status": "COMPLETE_FAILURE",
                "verified_count": 0,
                "failed_urls": [c["url"] for c in citations]
            }

    async def _recover_from_verification_failure(self, failed_urls: List[str], domain: str) -> List[Dict]:
        """Recover from verification failures using fallback sources"""

        # This would use the actual error recovery system
        # For this example, we'll return some fallback citations

        fallback_citations = [
            {
                "id": "general-mdn-web-docs",
                "title": "MDN Web Documentation",
                "url": "https://developer.mozilla.org/",
                "description": "Comprehensive web development documentation",
                "credibility": 10,
                "type": "official_docs"
            }
        ]

        return fallback_citations

    def _format_verified_citations(self, citations: List[Dict]) -> str:
        """Format verified citations in Korean style"""

        if not citations:
            return ""

        citations_text = "## 인용문\n\n"

        for i, citation in enumerate(citations, 1):
            credibility_icon = "🔵 공식 문서" if citation["credibility"] >= 10 else "🟢 신뢰 출처"

            citations_text += f"{i}. **{citation['title']}** {credibility_icon}\n"
            citations_text += f"   {citation['description']}\n"
            citations_text += f"   [{citation['url']}]({citation['url']})\n\n"

        citations_text += """**검증 기준**:
- ✅ 공식 문서 및 신뢰할 수 있는 출처만 사용
- ✅ 실시간 URL 접속 가능성 확인
- ✅ 콘텐츠 관련성 검증 완료
- ❌ 할루시네이션 방지를 위해 미확인 출처 제외"""

        return citations_text

    def _generate_citation_preparation_message(self) -> str:
        """Generate message when citations are being prepared"""

        return """## 인용문

📋 **인용문 준비 중**

현재 이 주제에 대한 신뢰할 수 있는 공식 출처를 조사 중입니다.
할루시네이션을 방지하기 위해 실시간 URL 검증을 수행한 후,
100% 검증된 출처만 인용문으로 추가하겠습니다.

**검증 기준**:
- ✅ 공식 문서 (Anthropic, Python.org, 등)
- ✅ 실시간 URL 접속 가능성 확인
- ✅ 콘텐츠 관련성 검증
- ❌ 가상 출처 또는 미확인 블로그 제외"""

    def _combine_content_and_citations(self, content: str, citations_section: str) -> str:
        """Combine chapter content with citation section"""

        return f"{content}\n\n{citations_section}"

    async def _final_validation(self, final_chapter: str, citations: List[Dict]) -> Dict[str, Any]:
        """Perform final validation on the complete chapter"""

        # This would implement comprehensive validation
        # For this example, we'll simulate basic validation

        import time
        start_time = time.time()

        # Simulate validation checks
        content_length = len(final_chapter)
        has_citations = "인용문" in final_chapter
        warnings = []

        if content_length < 1000:
            warnings.append("챕터 내용이 너무 짧습니다")

        if not citations and not has_citations:
            warnings.append("인용문이 없습니다")

        processing_time = time.time() - start_time

        return {
            "passed": len(warnings) == 0,
            "processing_time": processing_time,
            "warnings": warnings,
            "content_length": content_length,
            "has_citations": has_citations
        }


# Example usage
async def main():
    """Example of using the integrated system"""

    # Initialize the enhanced book author
    book_author = YodaBookAuthorWithCitationVerification()

    # Create a chapter generation request
    request = ChapterGenerationRequest(
        topic="Claude Code 기본 기능",
        chapter_title="Claude Code와의 첫 만남",
        chapter_num=1,
        target_audience="초급 개발자",
        domain="claude_code"
    )

    # Generate chapter with verified citations
    print("🚀 YodA Book Author with Citation Verification")
    print("=" * 50)

    result = await book_author.generate_chapter_with_verified_citations(request)

    # Display results
    print(f"\n📊 Generation Results:")
    print(f"✅ Status: {result['verification_status']}")
    print(f"📋 Citations: {result['citations_count']}")
    print(f"🕐 Processing Time: {result['processing_time']:.2f}s")
    print(f"✨ Validation Passed: {result['validation_passed']}")

    if result['warnings']:
        print(f"⚠️ Warnings: {len(result['warnings'])}")
        for warning in result['warnings']:
            print(f"   - {warning}")

    # Show a preview of the generated content
    print(f"\n📝 Content Preview:")
    print("-" * 30)
    print(result['chapter_content'][:500] + "..." if len(result['chapter_content']) > 500 else result['chapter_content'])


if __name__ == "__main__":
    asyncio.run(main())