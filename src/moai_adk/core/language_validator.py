"""
Language Validator

Provides comprehensive language validation capabilities for programming languages.

"""

from pathlib import Path
from typing import Any, Dict, List, Optional, Set, Tuple


# Language detector functionality removed due to missing dependency
# Using simplified language detection for now
def get_all_supported_languages():
    """Get all supported programming languages."""
    return {"python", "javascript", "typescript", "java", "go", "rust", "cpp", "c"}


def get_language_by_file_extension(extension: str) -> Optional[str]:
    """Get programming language by file extension."""

    # Handle Path objects and strings
    if hasattr(extension, "suffix"):
        # Path object
        ext = extension.suffix.lower()
    else:
        # String - extract extension
        ext = str(extension).lower()
        if not ext.startswith("."):
            # Extract extension from filename
            if "." in ext:
                ext = "." + ext.split(".")[-1]
            else:
                ext = ""

    EXTENSION_MAP = {  # noqa: N806 - intentional constant naming
        ".py": "python",
        ".js": "javascript",
        ".ts": "typescript",
        ".java": "java",
        ".go": "go",
        ".rs": "rust",
        ".cpp": "cpp",
        ".c": "c",
        ".pyw": "python",
        ".pyx": "python",
    }
    return EXTENSION_MAP.get(ext)


def is_code_directory(path: str) -> bool:
    """Check if directory is a code directory."""
    code_dirs = {"src", "lib", "app", "components", "modules", "packages"}
    return any(dir_name in path for dir_name in code_dirs)


LANGUAGE_DIRECTORY_MAP = {
    "python": ["src", "tests", "examples"],
    "javascript": ["src", "lib", "packages"],
    "typescript": ["src", "lib", "packages"],
}


def get_exclude_patterns():
    """Get patterns to exclude from language detection."""
    return ["*.pyc", "*.pyo", "__pycache__", ".git", "node_modules", ".venv"]


class LanguageValidator:
    """
    A comprehensive language validator for programming languages.

    This class provides language detection, validation, and project structure
    analysis capabilities based on existing language detection infrastructure.
    """

    # Extended file extension mapping for better language detection
    EXTENSION_MAP = {
        "python": [".py", ".pyw", ".pyx", ".pxd"],
        "javascript": [".js", ".jsx", ".mjs"],
        "typescript": [".ts", ".tsx", ".cts", ".mts"],
        "go": [".go"],
        "rust": [".rs"],
        "kotlin": [".kt", ".kts"],
        "ruby": [".rb"],
        "php": [".php", ".php3", ".php4", ".php5", ".phtml"],
        "java": [".java"],
        "csharp": [".cs"],
        "cpp": [".cpp", ".cxx", ".cc", ".c++", ".h", ".hpp"],
        "c": [".c", ".h"],
        "swift": [".swift"],
        "dart": [".dart"],
        "scala": [".scala"],
        "clojure": [".clj", ".cljs", ".cljc"],
        "haskell": [".hs", ".lhs"],
        "lua": [".lua"],
        "ocaml": [".ml", ".mli", ".mll", ".mly"],
        "elixir": [".ex", ".exs"],
        "bash": [".sh", ".bash"],
        "powershell": [".ps1", ".psm1", ".psd1"],
        "sql": [".sql"],
        "html": [".html", ".htm"],
        "css": [".css", ".scss", ".sass"],
        "json": [".json", ".json5"],
        "yaml": [".yaml", ".yml"],
        "toml": [".toml"],
        "xml": [".xml", ".xsl", ".xslt"],
        "markdown": [".md", ".markdown"],
        "dockerfile": ["dockerfile", "dockerfile.", "dockerfile.*"],
    }

    def __init__(
        self,
        supported_languages: Optional[List[str]] = None,
        auto_validate: bool = True,
    ):
        """
        Initialize the language validator.

        Args:
            supported_languages: List of supported language codes.
                                If None, uses all available languages.
            auto_validate: Whether to automatically validate inputs and perform cleanup
        """
        self.auto_validate = auto_validate

        if supported_languages is None:
            # Use all languages from the existing language detection system
            self.supported_languages = set(get_all_supported_languages())
        else:
            self.supported_languages = set(lang.lower() for lang in supported_languages)

        # Compile regex patterns for efficient matching
        self._directory_patterns: Dict[str, Any] = {}
        self._exclude_patterns_cache: Dict[str, Any] = {}

        # Initialize analysis cache for statistics tracking
        self._analysis_cache = {
            "last_analysis_files": 0,
            "detected_extensions": [],
            "supported_languages_found": 0,
        }

    def _validate_and_normalize_input(self, value: Any, input_type: str) -> Optional[Any]:
        """
        Validate and normalize input values.

        Args:
            value: Input value to validate
            input_type: Type of input ('language', 'file_path', 'list', etc.)

        Returns:
            Normalized value or None if validation fails
        """
        if not value and input_type != "language":  # Empty language is valid sometimes
            return None

        if input_type == "language":
            if not isinstance(value, str):
                return None
            return value.strip().lower() if value else None

        elif input_type == "file_path":
            if isinstance(value, str):
                return Path(value).resolve()
            elif isinstance(value, Path):
                return value.resolve()
            else:
                return None

        elif input_type == "list":
            if not isinstance(value, list):
                return None
            return value

        return None

    def validate_language(self, language: str) -> bool:
        """
        Validate if a language is supported.

        Args:
            language: Language code to validate.

        Returns:
            True if language is supported, False otherwise.
        """
        if self.auto_validate:
            normalized_lang = self._validate_and_normalize_input(language, "language")
            if normalized_lang is None:
                return False
        else:
            normalized_lang = self.normalize_language_code(language)

        return normalized_lang in self.supported_languages

    def detect_language_from_extension(self, file_path: Any) -> Optional[str]:
        """
        Detect language from file extension using enhanced mapping.

        Args:
            file_path: File path as string or Path object.

        Returns:
            Detected language code or None if not recognized.
        """
        if self.auto_validate:
            path_obj = self._validate_and_normalize_input(file_path, "file_path")
            if path_obj is None:
                return None
        else:
            if isinstance(file_path, str):
                path_obj = Path(file_path)
            elif isinstance(file_path, Path):
                path_obj = file_path
            else:
                return None

        # First try the enhanced mapping
        extension = path_obj.suffix.lower()
        for lang, extensions in self.EXTENSION_MAP.items():
            if extension in extensions:
                return lang

        # Fall back to existing system for backwards compatibility
        return get_language_by_file_extension(path_obj)

    def get_expected_directories(self, language: str) -> List[str]:
        """
        Get expected directory patterns for a language.

        Args:
            language: Language code.

        Returns:
            List of expected directory patterns.
        """
        if self.auto_validate:
            normalized_lang = self._validate_and_normalize_input(language, "language")
            if normalized_lang is None:
                return []
        else:
            normalized_lang = self.normalize_language_code(language)

        if normalized_lang in LANGUAGE_DIRECTORY_MAP:
            dirs = LANGUAGE_DIRECTORY_MAP[normalized_lang].copy()
            # Add trailing slash for consistency with test expectations
            return [f"{dir}/" if not dir.endswith("/") else dir for dir in dirs]

        # Return default Python directories as fallback with trailing slashes
        default_dirs = LANGUAGE_DIRECTORY_MAP.get("python", [])
        return [f"{dir}/" if not dir.endswith("/") else dir for dir in default_dirs]

    def get_file_extensions(self, language: str) -> List[str]:
        """
        Get file extensions for a language.

        Args:
            language: Language code.

        Returns:
            List of file extensions (including dot).
        """
        if self.auto_validate:
            normalized_lang = self._validate_and_normalize_input(language, "language")
            if normalized_lang is None:
                return []
        else:
            normalized_lang = self.normalize_language_code(language)

        return self.EXTENSION_MAP.get(normalized_lang, [])

    def get_all_supported_extensions(self) -> Set[str]:
        """
        Get all supported file extensions.

        Returns:
            Set of all supported file extensions.
        """
        all_extensions = set()
        for extensions in self.EXTENSION_MAP.values():
            all_extensions.update(extensions)
        return all_extensions

    def detect_language_from_filename(self, file_name: str) -> Optional[str]:
        """
        Detect language from filename (including special cases like Dockerfile).

        Args:
            file_name: Filename or full path.

        Returns:
            Detected language code or None if not recognized.
        """
        if self.auto_validate:
            normalized_name = self._validate_and_normalize_input(file_name, "file_path")
            if normalized_name is None:
                return None
        else:
            if not file_name or not isinstance(file_name, str):
                return None
            normalized_name = Path(file_name)

        # Extract filename from path if needed
        filename = normalized_name.name.lower()

        # Check for special filenames
        if filename in ["dockerfile", "dockerfile.dev", "dockerfile.prod"]:
            return "dockerfile"

        # Check for common build/config files
        config_patterns = {
            "makefile": "bash",
            "cmakelists.txt": "cpp",
            "pom.xml": "java",
            "build.gradle": "kotlin",
            "package.json": "javascript",
            "pyproject.toml": "python",
            "cargo.toml": "rust",
            "go.mod": "go",
            "requirements.txt": "python",
            "gemfile": "ruby",
        }

        if filename in config_patterns:
            return config_patterns[filename]

        # Extract extension and try normal detection
        Path(filename).suffix.lower()
        return self.detect_language_from_extension(filename)

    def validate_file_extension(self, file_path: Any, language: str) -> bool:
        """
        Validate if a file has the correct extension for a language.

        Args:
            file_path: File path to validate.
            language: Expected language code.

        Returns:
            True if file extension matches language, False otherwise.
        """
        if language is None:
            # Any file is valid when no specific language is required
            return True

        if self.auto_validate:
            normalized_lang = self._validate_and_normalize_input(language, "language")
            if normalized_lang is None:
                return False
        else:
            normalized_lang = self.normalize_language_code(language)

        detected_lang = self.detect_language_from_extension(file_path)
        return detected_lang == normalized_lang

    def get_supported_languages(self) -> List[str]:
        """
        Get list of supported languages.

        Returns:
            Sorted list of supported language codes.
        """
        return sorted(self.supported_languages)

    def normalize_language_code(self, language: str) -> str:
        """
        Normalize language code to lowercase with stripped whitespace.

        Args:
            language: Raw language code.

        Returns:
            Normalized language code.
        """
        if not language or not isinstance(language, str):
            return ""

        return language.strip().lower()

    def validate_project_configuration(self, config: Dict[str, Any]) -> Tuple[bool, List[str]]:
        """
        Validate project configuration for language support.

        Args:
            config: Project configuration dictionary.

        Returns:
            Tuple of (is_valid, issues) where is_valid is boolean and issues is list of strings.
        """
        if self.auto_validate:
            validated_config = self._validate_and_normalize_input(config, "dict")
            if validated_config is None:
                return False, ["Invalid configuration format"]

        issues = []

        # Check if project section exists
        if "project" not in config:
            issues.append("Missing 'project' section in configuration")
            return False, issues

        project_config = config["project"]

        # Check if language is specified
        if "language" not in project_config:
            issues.append("Missing 'language' field in project configuration")
            return False, issues

        project_language = project_config["language"]

        # Validate the language
        if not self.validate_language(project_language):
            issues.append(f"Unsupported language: {project_language}")
            return False, issues

        # Check if name is specified
        if "name" not in project_config:
            issues.append("Missing 'name' field in project configuration")
            return False, issues

        # Check if name is valid (not empty)
        if not project_config["name"] or not isinstance(project_config["name"], str):
            issues.append("Project name must be a non-empty string")
            return False, issues

        # Additional validation for empty strings and whitespace-only names
        if isinstance(project_config["name"], str) and not project_config["name"].strip():
            issues.append("Project name cannot be empty or contain only whitespace")
            return False, issues

        return True, issues

    def validate_project_structure(self, project_files: Dict[str, bool], language: str) -> Tuple[bool, List[str]]:
        """
        Validate project structure for a specific language.

        Args:
            project_files: Dictionary mapping file paths to boolean (is_source_file).
            language: Project language to validate against.

        Returns:
            Tuple of (is_valid, issues) where is_valid is boolean and issues is list of strings.
        """
        if self.auto_validate:
            validated_project_files = self._validate_and_normalize_input(project_files, "dict")
            validated_language = self._validate_and_normalize_input(language, "language")
            if validated_project_files is None or validated_language is None:
                return False, ["Invalid input format for project structure validation"]

        issues = []
        expected_dirs = self.get_expected_directories(language)

        # Group files by directory
        files_by_dir: Dict[str, List[str]] = {}
        for file_path, is_source in project_files.items():
            if is_source:  # Only validate source files
                dir_path = str(Path(file_path).parent) + "/"
                if dir_path not in files_by_dir:
                    files_by_dir[dir_path] = []
                files_by_dir[dir_path].append(file_path)

        # Check if expected directories exist and have files
        for expected_dir in expected_dirs:
            found_files_in_dir = False
            for actual_dir in files_by_dir:
                if actual_dir.startswith(expected_dir):
                    found_files_in_dir = True
                    break

            if not found_files_in_dir and expected_dir != "{package_name}/":
                issues.append(f"No source files found in expected directory: {expected_dir}")

        # Check for files in unexpected directories
        # Note: Using simplified check since is_code_directory signature changed
        for file_path, is_source in project_files.items():
            if is_source:
                path_str = str(file_path)
                if not is_code_directory(path_str):
                    issues.append(f"Source file in unexpected location: {file_path}")

        return len(issues) == 0, issues

    def get_language_statistics(self, files: List[Any]) -> Dict[str, int]:
        """
        Get language statistics from a list of files.

        Args:
            files: List of file paths.

        Returns:
            Dictionary mapping language codes to file counts.
        """
        if self.auto_validate:
            validated_files = self._validate_and_normalize_input(files, "list")
            if validated_files is None:
                return {}
        else:
            validated_files = files

        stats: Dict[str, int] = {}
        total_files = 0
        detected_extensions = set()

        for file_path in validated_files:
            if file_path:  # Ensure file path is not None
                detected_lang = self.detect_language_from_extension(file_path)
                if detected_lang:
                    stats[detected_lang] = stats.get(detected_lang, 0) + 1
                    total_files += 1

                    # Track detected extensions for analysis
                    if hasattr(file_path, "suffix"):
                        detected_extensions.add(file_path.suffix.lower())
                    elif isinstance(file_path, str):
                        detected_extensions.add(Path(file_path).suffix.lower())

        # Add analysis information
        if hasattr(self, "_analysis_cache"):
            self._analysis_cache["last_analysis_files"] = total_files
            self._analysis_cache["detected_extensions"] = list(detected_extensions)
            self._analysis_cache["supported_languages_found"] = len(stats)

        return stats

    def get_analysis_cache(self) -> Dict[str, Any]:
        """
        Get the analysis cache with language detection statistics.

        Returns:
            Dictionary containing analysis statistics.
        """
        return self._analysis_cache.copy()

    def clear_analysis_cache(self) -> None:
        """
        Clear the analysis cache.
        """
        self._analysis_cache = {
            "last_analysis_files": 0,
            "detected_extensions": [],
            "supported_languages_found": 0,
        }
