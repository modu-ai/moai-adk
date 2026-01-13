"""TAG System v2.0 for MoAI-ADK.

Provides TAG annotation, validation, parsing, and linkage management
for SPEC-First TDD workflow.
"""

from .linkage import (
    LinkageManager,
    spec_document_exists,
)
from .parser import (
    extract_tags_from_directory,
    extract_tags_from_file,
    extract_tags_from_files,
    extract_tags_from_source,
)
from .validator import (
    DEFAULT_VERB,
    TAG,
    VALID_VERBS,
    get_default_verb,
    parse_tag_string,
    validate_spec_id_format,
    validate_tag,
    validate_verb,
)

__all__ = [
    # Validator
    "TAG",
    "VALID_VERBS",
    "DEFAULT_VERB",
    "validate_spec_id_format",
    "validate_verb",
    "get_default_verb",
    "validate_tag",
    "parse_tag_string",
    # Parser
    "extract_tags_from_source",
    "extract_tags_from_file",
    "extract_tags_from_files",
    "extract_tags_from_directory",
    # Linkage
    "LinkageManager",
    "spec_document_exists",
]

__version__ = "2.0.0"
