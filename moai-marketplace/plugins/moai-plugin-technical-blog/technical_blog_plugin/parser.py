"""
Directive Parser: Automatically detects blog writing mode and template selection
Handles natural language input from /blog-write command
"""

import re
from enum import Enum
from dataclasses import dataclass
from typing import Optional


class Mode(Enum):
    """Blog writing modes"""
    CREATE = "create"      # Write new blog post
    OPTIMIZE = "optimize"  # Optimize existing post
    LIST = "list"         # List templates


@dataclass
class ParsedDirective:
    """Parsed directive result"""
    mode: Mode
    template_id: Optional[str] = None
    topic: Optional[str] = None
    file_path: Optional[str] = None
    difficulty: Optional[str] = None
    raw_directive: str = ""


class DirectiveParser:
    """Parse /blog-write natural language directives into structured commands"""

    # Template selection keywords
    TEMPLATE_KEYWORDS = {
        "tutorial": ["튜토리얼", "가이드", "배우기", "시작하기", "tutorial", "guide"],
        "case-study": ["케이스 스터디", "사례", "성공", "개선", "마이그레이션", "case study", "migration"],
        "howto": ["방법", "어떻게", "구현", "how to", "implement"],
        "announcement": ["발표", "공지", "소개", "릴리즈", "announce", "release"],
        "comparison": ["비교", "vs", "차이점", "comparison", "vs "],
    }

    # List mode keywords
    LIST_KEYWORDS = ["템플릿", "템플릿 목록", "list", "templates"]

    # Difficulty keywords
    DIFFICULTY_KEYWORDS = {
        "beginner": ["초보자", "beginner", "입문", "기초"],
        "intermediate": ["중급", "intermediate", "일반"],
        "advanced": ["고급", "advanced", "심화"],
    }

    @staticmethod
    def parse(directive: str) -> ParsedDirective:
        """
        Parse a blog-write directive and return structured command

        Args:
            directive: Natural language directive from /blog-write

        Returns:
            ParsedDirective with mode, template, topic, etc.
        """
        directive_lower = directive.lower()

        # Check for LIST mode (template listing)
        if DirectiveParser._is_list_mode(directive_lower):
            return ParsedDirective(
                mode=Mode.LIST,
                raw_directive=directive
            )

        # Check for OPTIMIZE mode (existing file optimization)
        if DirectiveParser._is_optimize_mode(directive, directive_lower):
            file_path = DirectiveParser._extract_file_path(directive)
            return ParsedDirective(
                mode=Mode.OPTIMIZE,
                file_path=file_path,
                raw_directive=directive
            )

        # Default: CREATE mode (new post)
        template_id = DirectiveParser._detect_template(directive_lower)
        difficulty = DirectiveParser._detect_difficulty(directive_lower)
        topic = DirectiveParser._extract_topic(directive, template_id)

        return ParsedDirective(
            mode=Mode.CREATE,
            template_id=template_id or "tutorial",  # Default to tutorial
            topic=topic,
            difficulty=difficulty,
            raw_directive=directive
        )

    @staticmethod
    def _is_list_mode(directive_lower: str) -> bool:
        """Check if directive is requesting template list"""
        return any(keyword in directive_lower for keyword in DirectiveParser.LIST_KEYWORDS)

    @staticmethod
    def _is_optimize_mode(directive: str, directive_lower: str) -> bool:
        """Check if directive contains file path for optimization"""
        optimize_keywords = ["최적화", "개선", "optimize", "improve"]
        has_optimize_keyword = any(kw in directive_lower for kw in optimize_keywords)
        has_file_path = ".md" in directive or "./posts/" in directive
        return has_optimize_keyword and has_file_path

    @staticmethod
    def _detect_template(directive_lower: str) -> Optional[str]:
        """Detect which template to use based on keywords"""
        for template_id, keywords in DirectiveParser.TEMPLATE_KEYWORDS.items():
            for keyword in keywords:
                if keyword in directive_lower:
                    return template_id
        return None

    @staticmethod
    def _detect_difficulty(directive_lower: str) -> Optional[str]:
        """Detect difficulty level from keywords"""
        for difficulty, keywords in DirectiveParser.DIFFICULTY_KEYWORDS.items():
            for keyword in keywords:
                if keyword in directive_lower:
                    return difficulty
        return None

    @staticmethod
    def _extract_file_path(directive: str) -> Optional[str]:
        """Extract file path from directive"""
        # Look for patterns like "./posts/filename.md" or "filename.md"
        match = re.search(r'([./a-zA-Z0-9_-]+\.md)', directive)
        if match:
            return match.group(1)
        return None

    @staticmethod
    def _extract_topic(directive: str, template_id: Optional[str]) -> Optional[str]:
        """Extract topic/title from directive"""
        # Remove common action verbs and keywords
        remove_patterns = [
            r"작성$",  # Korean: "작성" (write)
            r"튜토리얼.*$",  # Remove template keywords
            r"케이스 스터디.*$",
            r"가이드.*$",
            r"비교.*$",
            r"발표.*$",
            r"작성해$",
            r"\/blog-write\s*",  # Remove command prefix
            r"\"",  # Remove quotes
            r"'",
        ]

        topic = directive
        for pattern in remove_patterns:
            topic = re.sub(pattern, "", topic)

        return topic.strip() if topic.strip() else None


# Example usage
if __name__ == "__main__":
    # Test cases
    test_directives = [
        "Next.js 15 초보자 튜토리얼 작성",
        "마이그레이션으로 20% 성능 향상한 사례 작성",
        "./posts/nextjs-tutorial.md 최적화",
        "템플릿 목록",
        "React vs Vue 비교 분석",
    ]

    parser = DirectiveParser()
    for directive in test_directives:
        result = parser.parse(directive)
        print(f"Input: {directive}")
        print(f"  Mode: {result.mode.value}")
        print(f"  Template: {result.template_id}")
        print(f"  Topic: {result.topic}")
        print(f"  Difficulty: {result.difficulty}")
        print()
