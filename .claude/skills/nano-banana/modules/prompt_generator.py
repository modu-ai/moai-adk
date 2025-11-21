"""
Nano Banana Pro Prompt Generation Module

Convert natural language requests into optimized Nano Banana Pro prompts.

Usage:
    from modules.prompt_generator import PromptGenerator

    # Optimize prompt
    optimized_prompt = PromptGenerator.optimize("beautiful Korean mountain landscape")

    # Validate prompt
    is_valid = PromptGenerator.validate("mountain landscape")
"""

from typing import Optional, Dict
import re


class PromptGenerator:
    """Nano Banana Pro Prompt Generation and Optimization Class"""

    # Nano Banana Pro prompt constraints
    MAX_TOKENS = 1000
    MAX_CHARS = 2000

    # Photographic enhancement keywords (SynthID evasion prevention)
    PHOTOGRAPHIC_ELEMENTS = [
        "professional photography",
        "highly detailed",
        "sharp focus",
        "cinematic lighting",
        "volumetric lighting",
        "shot on 50mm lens",
        "f/2.8",
        "award-winning",
        "best quality",
        "masterpiece",
    ]

    # Resolution presets
    RESOLUTION_PRESETS = {
        "1k": "1024x1024",
        "2k": "2048x2048",
        "4k": "4096x4096",
    }

    # Style templates
    STYLE_TEMPLATES = {
        "photorealistic": "photorealistic, hyper-realistic, professional photography",
        "artistic": "artistic, painterly, illustration style",
        "cinematic": "cinematic, movie-like, dramatic lighting",
        "minimal": "minimalist, clean, simple composition",
        "abstract": "abstract, geometric, modern art",
        "fantasy": "fantasy art, magical, ethereal",
    }

    @classmethod
    def optimize(
        cls,
        prompt: str,
        style: Optional[str] = None,
        add_photographic: bool = True,
        language: str = "auto",
    ) -> str:
        """
        Convert natural language prompt to optimized Nano Banana Pro prompt.

        Args:
            prompt: Original prompt
            style: Style (photorealistic, artistic, cinematic, etc.)
            add_photographic: Whether to add photographic elements
            language: Prompt language (auto, ko, en, ja, zh)

        Returns:
            str: Optimized prompt
        """
        if not cls.validate(prompt):
            raise ValueError(f"Invalid prompt: {prompt}")

        # Clean prompt
        cleaned = cls._clean_prompt(prompt)

        # Add style
        if style and style in cls.STYLE_TEMPLATES:
            cleaned = f"{cleaned}, {cls.STYLE_TEMPLATES[style]}"

        # Add photographic elements
        if add_photographic:
            photographic = ", ".join(cls.PHOTOGRAPHIC_ELEMENTS[:5])
            cleaned = f"{cleaned}, {photographic}"

        # Language-specific optimization
        cleaned = cls._language_optimize(cleaned, language)

        # Apply length constraints
        return cls._truncate(cleaned, cls.MAX_CHARS)

    @classmethod
    def _clean_prompt(cls, prompt: str) -> str:
        """
        Clean prompt text.

        Args:
            prompt: Original prompt

        Returns:
            str: Cleaned prompt
        """
        # Clean multiple lines
        cleaned = " ".join(prompt.split())

        # Clean special characters (except comma, period, etc.)
        cleaned = re.sub(r"[^\w\s\.,!?가-힣a-zA-Z]", "", cleaned)

        # Remove duplicate spaces
        cleaned = re.sub(r"\s+", " ", cleaned)

        return cleaned.strip()

    @classmethod
    def _language_optimize(cls, prompt: str, language: str) -> str:
        """
        Apply language-specific prompt optimization.

        Args:
            prompt: Original prompt
            language: Target language

        Returns:
            str: Optimized prompt
        """
        language = language.lower()

        # Auto-detect language
        if language == "auto":
            language = cls._detect_language(prompt)

        # Language-specific optimization
        optimization_map = {
            "ko": ", Korean style, traditional elements",
            "ja": ", Japanese aesthetic, traditional art",
            "zh": ", Chinese landscape painting style",
            "en": ", Western style, contemporary",
        }

        if language in optimization_map:
            prompt = prompt + optimization_map[language]

        return prompt

    @classmethod
    def _detect_language(cls, text: str) -> str:
        """
        Detect language of text.

        Args:
            text: Text to analyze

        Returns:
            str: Detected language code
        """
        # Korean range: U+AC00 ~ U+D7AF
        if re.search(r"[가-힣]", text):
            return "ko"
        # Japanese range: Hiragana, Katakana, Kanji
        elif re.search(r"[\u3040-\u309F\u30A0-\u30FF]", text):
            return "ja"
        # Chinese range
        elif re.search(r"[\u4E00-\u9FFF]", text):
            return "zh"
        else:
            return "en"

    @classmethod
    def _truncate(cls, text: str, max_length: int) -> str:
        """
        Truncate text to maximum length.

        Args:
            text: Original text
            max_length: Maximum length

        Returns:
            str: Truncated text
        """
        if len(text) <= max_length:
            return text

        # Cut at last comma or space
        truncated = text[:max_length].rsplit(",", 1)[0].strip()
        return truncated if truncated else text[:max_length]

    @classmethod
    def validate(cls, prompt: str) -> bool:
        """
        Validate prompt format.

        Args:
            prompt: Prompt to validate

        Returns:
            bool: Validation status
        """
        if not prompt or not isinstance(prompt, str):
            return False

        if len(prompt) < 3 or len(prompt) > cls.MAX_CHARS:
            return False

        return True

    @classmethod
    def add_style(cls, prompt: str, style: str) -> str:
        """
        Add style to prompt.

        Args:
            prompt: Original prompt
            style: Style to add

        Returns:
            str: Prompt with style added
        """
        if style not in cls.STYLE_TEMPLATES:
            return prompt

        return f"{prompt}, {cls.STYLE_TEMPLATES[style]}"

    @classmethod
    def get_style_list(cls) -> list:
        """
        Return list of available styles.

        Returns:
            list: List of styles
        """
        return list(cls.STYLE_TEMPLATES.keys())

    @classmethod
    def get_resolution_list(cls) -> Dict[str, str]:
        """
        Return list of available resolutions.

        Returns:
            dict: Resolution mapping
        """
        return cls.RESOLUTION_PRESETS.copy()


if __name__ == "__main__":
    # Test execution
    prompt1 = PromptGenerator.optimize(
        "beautiful Korean mountain landscape",
        style="photorealistic"
    )
    print(f"Optimized (Korean): {prompt1}\n")

    prompt2 = PromptGenerator.optimize(
        "peaceful ocean sunset",
        style="cinematic"
    )
    print(f"Optimized (English): {prompt2}\n")

    print(f"Available styles: {PromptGenerator.get_style_list()}")
