"""Config API Models

Pydantic models for configuration API endpoints.
Defines request/response schemas for configuration management.
"""

from typing import Any, Dict, List, Optional, Union

from pydantic import BaseModel, Field

# ===============================
# Schema Models
# ===============================


class QuestionOption(BaseModel):
    """Option for a select question."""

    label: str
    value: Union[str, bool, int]
    description: Optional[str] = None


class Question(BaseModel):
    """A single question in a batch."""

    id: str
    question: str
    type: str = Field(..., pattern="^(text_input|select_single|number_input)$")
    required: bool = True
    options: List[QuestionOption] = Field(default_factory=list)
    smart_default: Optional[Union[str, bool, int]] = None
    min: Optional[int] = None
    max: Optional[int] = None
    show_if: Optional[str] = None
    conditional_mapping: Optional[Dict[str, List[str]]] = None
    smart_default_mapping: Optional[Dict[str, Union[str, bool]]] = None


class Batch(BaseModel):
    """A batch of questions in a tab."""

    id: str
    header: str
    batch_number: int
    total_batches: int
    questions: List[Question] = Field(default_factory=list)
    show_if: Optional[str] = None


class Tab(BaseModel):
    """A tab in the configuration schema."""

    id: str
    label: str
    description: str
    batches: List[Batch] = Field(default_factory=list)


class TabSchemaResponse(BaseModel):
    """Response model for tab schema endpoint."""

    version: str = "3.0.0"
    tabs: List[Tab]


# ===============================
# Config Models
# ===============================


class UserConfig(BaseModel):
    """User configuration section."""

    name: str


class LanguageConfig(BaseModel):
    """Language configuration section."""

    conversation_language: str
    agent_prompt_language: str = "en"
    conversation_language_name: Optional[str] = None


class ProjectConfig(BaseModel):
    """Project configuration section."""

    name: str
    description: str = ""
    language: str = ""
    locale: str = ""
    template_version: str = ""
    documentation_mode: str = "skip"
    documentation_depth: Optional[str] = None


class GitHubConfig(BaseModel):
    """GitHub configuration section."""

    profile_name: Optional[str] = None


class GitStrategyPersonalConfig(BaseModel):
    """Personal git strategy configuration."""

    workflow: str = "github-flow"
    auto_checkpoint: str = "disabled"
    push_to_remote: bool = False


class GitStrategyTeamConfig(BaseModel):
    """Team git strategy configuration."""

    workflow: str = "git-flow"
    auto_pr: bool = False
    draft_pr: bool = False


class GitStrategyConfig(BaseModel):
    """Git strategy configuration section."""

    mode: str = "manual"
    workflow: Optional[str] = None
    personal: GitStrategyPersonalConfig = Field(default_factory=GitStrategyPersonalConfig)
    team: GitStrategyTeamConfig = Field(default_factory=GitStrategyTeamConfig)


class ConstitutionConfig(BaseModel):
    """Constitution/quality configuration section."""

    test_coverage_target: int = Field(default=85, ge=0, le=100)
    enforce_tdd: bool = True


class MoAIConfig(BaseModel):
    """MoAI framework configuration section."""

    version: str = ""


class FullConfig(BaseModel):
    """Complete configuration model."""

    version: str = "3.0.0"
    user: Optional[UserConfig] = None
    language: Optional[LanguageConfig] = None
    project: Optional[ProjectConfig] = None
    github: Optional[GitHubConfig] = None
    git_strategy: Optional[GitStrategyConfig] = None
    constitution: Optional[ConstitutionConfig] = None
    moai: Optional[MoAIConfig] = None


# ===============================
# Request/Response Models
# ===============================


class SaveConfigRequest(BaseModel):
    """Request model for saving configuration."""

    config: Dict[str, Any]


class SaveConfigResponse(BaseModel):
    """Response model for save configuration endpoint."""

    success: bool
    message: str = ""
    backup_path: Optional[str] = None


class ValidateConfigRequest(BaseModel):
    """Request model for validating configuration."""

    config: Dict[str, Any]


class ValidateConfigResponse(BaseModel):
    """Response model for validate configuration endpoint."""

    valid: bool
    missing_fields: List[str] = Field(default_factory=list)
    errors: List[str] = Field(default_factory=list)


# ===============================
# Flat Config Model
# ===============================


class FlatConfigValue(BaseModel):
    """A single flat configuration value."""

    key: str
    value: Any
    section: str
    type: str  # 'text', 'select', 'number', 'toggle'


class ConfigSection(BaseModel):
    """A configuration section for UI rendering."""

    id: str
    label: str
    fields: List[FlatConfigValue]
