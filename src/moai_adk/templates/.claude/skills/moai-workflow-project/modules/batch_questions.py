"""
Batch Questions Manager - Integrated Question System for MoAI Skills

Implements unified question collection, validation, and processing system
that consolidates question functionality from 5 different skills into one
cohesive module.

Author: MoAI-ADK
Version: 1.0.0
"""

import json
import logging
import os
from dataclasses import dataclass, field
from enum import Enum
from pathlib import Path
from typing import Any, Callable, Dict, List, Optional

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class QuestionType(Enum):
    """Question types supported by the system."""

    SINGLE_CHOICE = "single_choice"
    MULTI_CHOICE = "multi_choice"
    TEXT_INPUT = "text_input"
    BOOLEAN = "boolean"
    NUMBER_INPUT = "number_input"
    FILE_SELECT = "file_select"


class ValidationRule(Enum):
    """Validation rules for responses."""

    REQUIRED = "required"
    MIN_LENGTH = "min_length"
    MAX_LENGTH = "max_length"
    REGEX = "regex"
    CUSTOM = "custom"


@dataclass
class QuestionOption:
    """Represents a question option."""

    label: str
    value: str
    description: Optional[str] = None
    conditional_questions: Optional[List["Question"]] = None


@dataclass
class ValidationConfig:
    """Configuration for response validation."""

    rules: List[ValidationRule] = field(default_factory=list)
    parameters: Dict[str, Any] = field(default_factory=dict)
    error_message: Optional[str] = None
    custom_validator: Optional[Callable] = None


@dataclass
class Question:
    """Represents a single question in the batch."""

    id: str
    text: str
    type: QuestionType
    options: List[QuestionOption] = field(default_factory=list)
    validation: Optional[ValidationConfig] = None
    description: Optional[str] = None
    required: bool = True
    default_value: Any = None
    conditional_on: Optional[Dict[str, Any]] = None  # {"question_id": "expected_value"}
    metadata: Dict[str, Any] = field(default_factory=dict)


@dataclass
class QuestionBatch:
    """Represents a batch of questions."""

    id: str
    title: str
    description: str
    questions: List[Question] = field(default_factory=list)
    metadata: Dict[str, Any] = field(default_factory=dict)
    created_at: Optional[str] = None
    updated_at: Optional[str] = None


@dataclass
class UserResponse:
    """Represents a user's response to a question."""

    question_id: str
    value: Any
    metadata: Dict[str, Any] = field(default_factory=dict)


class BatchQuestionsManager:
    """
    Unified question management system for MoAI skills.

    Consolidates question functionality from:
    - moai-project-batch-questions
    - moai-project-config-manager
    - moai-project-language-initializer
    - moai-spec-intelligent-workflow
    - moai-menu-project
    """

    def __init__(self, config: Optional[Dict[str, Any]] = None):
        """Initialize the batch questions manager."""
        self.config = config or {}
        self.template_dir = Path(__file__).parent.parent / "templates" / "question-templates"
        self.cache_dir = Path(__file__).parent.parent / "cache"

        # Ensure directories exist
        self.template_dir.mkdir(parents=True, exist_ok=True)
        self.cache_dir.mkdir(parents=True, exist_ok=True)

        # Initialize template cache
        self._template_cache = {}

        # Load built-in templates
        self._load_builtin_templates()

    def create_batch(self, batch_config: Dict[str, Any]) -> QuestionBatch:
        """
        Create a question batch from configuration.

        Args:
            batch_config: Configuration dictionary for the batch

        Returns:
            QuestionBatch: Created question batch
        """
        questions = []

        for q_config in batch_config.get("questions", []):
            # Convert options
            options = []
            for opt_config in q_config.get("options", []):
                option = QuestionOption(
                    label=opt_config["label"],
                    value=opt_config["value"],
                    description=opt_config.get("description"),
                    conditional_questions=self._parse_conditional_questions(
                        opt_config.get("conditional_questions", [])
                    ),
                )
                options.append(option)

            # Convert validation
            validation = None
            if "validation" in q_config:
                val_config = q_config["validation"]
                validation = ValidationConfig(
                    rules=[ValidationRule(rule) for rule in val_config.get("rules", [])],
                    parameters=val_config.get("parameters", {}),
                    error_message=val_config.get("error_message"),
                    custom_validator=val_config.get("custom_validator"),
                )

            # Create question
            question = Question(
                id=q_config["id"],
                text=q_config["text"],
                type=QuestionType(q_config["type"]),
                options=options,
                validation=validation,
                description=q_config.get("description"),
                required=q_config.get("required", True),
                default_value=q_config.get("default_value"),
                conditional_on=q_config.get("conditional_on"),
                metadata=q_config.get("metadata", {}),
            )
            questions.append(question)

        batch = QuestionBatch(
            id=batch_config["id"],
            title=batch_config["title"],
            description=batch_config["description"],
            questions=questions,
            metadata=batch_config.get("metadata", {}),
            created_at=batch_config.get("created_at"),
            updated_at=batch_config.get("updated_at"),
        )

        return batch

    def load_template(self, template_name: str) -> Optional[QuestionBatch]:
        """
        Load a question template from file or cache.

        Args:
            template_name: Name of the template to load

        Returns:
            QuestionBatch: Loaded template or None if not found
        """
        # Check cache first
        if template_name in self._template_cache:
            return self._template_cache[template_name]

        # Load from file
        template_file = self.template_dir / f"{template_name}.json"
        if not template_file.exists():
            logger.warning(f"Template not found: {template_name}")
            return None

        try:
            with open(template_file, "r", encoding="utf-8") as f:
                template_config = json.load(f)

            batch = self.create_batch(template_config)

            # Cache the template
            self._template_cache[template_name] = batch

            return batch

        except Exception as e:
            logger.error(f"Error loading template {template_name}: {e}")
            return None

    def filter_questions(self, batch: QuestionBatch, context: Dict[str, Any]) -> List[Question]:
        """
        Filter questions based on context and conditional logic.

        Args:
            batch: Question batch to filter
            context: Current context (previous answers, config values, etc.)

        Returns:
            List[Question]: Filtered list of questions
        """
        filtered_questions = []

        for question in batch.questions:
            # Check conditional logic
            if question.conditional_on:
                should_show = self._evaluate_condition(question.conditional_on, context)
                if not should_show:
                    continue

            filtered_questions.append(question)

        return filtered_questions

    def execute_batch(self, batch: QuestionBatch, context: Optional[Dict[str, Any]] = None) -> Dict[str, UserResponse]:
        """
        Execute a question batch and collect responses.

        Args:
            batch: Question batch to execute
            context: Current context for conditional questions

        Returns:
            Dict[str, UserResponse]: Collected responses
        """
        if context is None:
            context = {}

        # Filter questions based on context
        questions = self.filter_questions(batch, context)

        responses = {}

        for question in questions:
            # Check if question already has response in context
            if question.id in context:
                responses[question.id] = UserResponse(
                    question_id=question.id,
                    value=context[question.id],
                    metadata={"source": "context"},
                )
                continue

            # Ask the question
            response_value = self._ask_question(question)

            # Validate response
            if question.validation:
                validation_result = self._validate_response(response_value, question.validation)
                if not validation_result.is_valid:
                    # Handle validation error
                    raise ValueError(f"Validation failed for question {question.id}: {validation_result.error}")

            responses[question.id] = UserResponse(
                question_id=question.id,
                value=response_value,
                metadata={"source": "user_input"},
            )

            # Update context for subsequent questions
            context[question.id] = response_value

        return responses

    def get_skill_specific_template(self, skill_name: str) -> Optional[QuestionBatch]:
        """
        Get a skill-specific question template.

        Args:
            skill_name: Name of the skill

        Returns:
            QuestionBatch: Skill-specific template or None
        """
        template_mapping = {
            "moai-project-batch-questions": "project-batch-questions",
            "moai-project-config-manager": "config-manager-setup",
            "moai-project-language-initializer": "language-initializer",
            "moai-spec-intelligent-workflow": "spec-workflow-setup",
            "moai-menu-project": "menu-project-config",
        }

        template_name = template_mapping.get(skill_name)
        if not template_name:
            logger.warning(f"No template mapping for skill: {skill_name}")
            return None

        return self.load_template(template_name)

    def process_responses_for_config(
        self, responses: Dict[str, UserResponse], config_path: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Process responses and generate configuration updates.

        Args:
            responses: User responses to process
            config_path: Path to existing config file (optional)

        Returns:
            Dict[str, Any]: Updated configuration
        """
        # Start with existing config if provided
        config = {}
        if config_path and os.path.exists(config_path):
            try:
                with open(config_path, "r", encoding="utf-8") as f:
                    config = json.load(f)
            except Exception as e:
                logger.error(f"Error loading existing config: {e}")

        # Process each response
        for response in responses.values():
            # Apply response to config based on metadata
            config_path = response.metadata.get("config_path", response.question_id)
            value = response.value

            # Handle special cases
            if isinstance(value, list) and len(value) == 1:
                # Single selection in multi-choice format
                value = value[0]

            # Nested path handling (e.g., "git_strategy.mode")
            if "." in config_path:
                self._set_nested_value(config, config_path, value)
            else:
                config[config_path] = value

        return config

    def save_responses(self, responses: Dict[str, UserResponse], file_path: str) -> bool:
        """
        Save responses to a file.

        Args:
            responses: Responses to save
            file_path: Path to save the responses

        Returns:
            bool: True if saved successfully
        """
        try:
            # Convert responses to serializable format
            serializable_responses = {}
            for resp in responses.values():
                serializable_responses[resp.question_id] = {
                    "value": resp.value,
                    "metadata": resp.metadata,
                }

            with open(file_path, "w", encoding="utf-8") as f:
                json.dump(serializable_responses, f, indent=2, ensure_ascii=False)

            return True

        except Exception as e:
            logger.error(f"Error saving responses: {e}")
            return False

    def _load_builtin_templates(self):
        """Load built-in question templates."""
        builtin_templates = {
            "project-initialization": self._create_project_init_template(),
            "git-workflow-config": self._create_git_workflow_template(),
            "quality-gates-setup": self._create_quality_gates_template(),
            "skill-selection": self._create_skill_selection_template(),
        }

        for name, template in builtin_templates.items():
            template_file = self.template_dir / f"{name}.json"
            if not template_file.exists():
                self._save_template_to_file(template, template_file)

    def _create_project_init_template(self) -> Dict[str, Any]:
        """Create project initialization template."""
        return {
            "id": "project-initialization",
            "title": "Project Initialization",
            "description": "Configure your new project settings",
            "questions": [
                {
                    "id": "project.name",
                    "text": "What is your project name?",
                    "type": "text_input",
                    "required": True,
                    "validation": {
                        "rules": ["required", "min_length"],
                        "parameters": {"min_length": 2},
                    },
                },
                {
                    "id": "project.type",
                    "text": "What type of project are you creating?",
                    "type": "single_choice",
                    "required": True,
                    "options": [
                        {"label": "Web Application", "value": "web"},
                        {"label": "API Service", "value": "api"},
                        {"label": "CLI Tool", "value": "cli"},
                        {"label": "Library", "value": "library"},
                    ],
                },
                {
                    "id": "project.language",
                    "text": "What programming language will you use?",
                    "type": "single_choice",
                    "required": True,
                    "options": [
                        {"label": "Python", "value": "python"},
                        {"label": "TypeScript", "value": "typescript"},
                        {"label": "JavaScript", "value": "javascript"},
                        {"label": "Go", "value": "go"},
                        {"label": "Rust", "value": "rust"},
                    ],
                },
            ],
        }

    def _create_git_workflow_template(self) -> Dict[str, Any]:
        """Create Git workflow configuration template."""
        return {
            "id": "git-workflow-config",
            "title": "Git Workflow Configuration",
            "description": "Configure your Git workflow preferences",
            "questions": [
                {
                    "id": "git_strategy.mode",
                    "text": "What Git workflow mode do you prefer?",
                    "type": "single_choice",
                    "required": True,
                    "options": [
                        {
                            "label": "Manual (local only)",
                            "value": "manual",
                            "description": "Work locally without automatic remote operations",
                        },
                        {
                            "label": "Personal (GitHub personal project)",
                            "value": "personal",
                            "description": "Automatic pushing to your personal GitHub repository",
                        },
                        {
                            "label": "Team (GitHub team project)",
                            "value": "team",
                            "description": "Team collaboration with pull requests and reviews",
                        },
                    ],
                },
                {
                    "id": "git_strategy.branch_creation.prompt_always",
                    "text": "Do you want to be prompted for branch creation?",
                    "type": "boolean",
                    "required": True,
                    "default_value": True,
                    "conditional_on": {"git_strategy.mode": ["personal", "team"]},
                },
                {
                    "id": "git_strategy.branch_creation.auto_enabled",
                    "text": "Enable automatic branch creation?",
                    "type": "boolean",
                    "required": True,
                    "default_value": False,
                    "conditional_on": {
                        "git_strategy.mode": ["personal", "team"],
                        "git_strategy.branch_creation.prompt_always": False,
                    },
                },
            ],
        }

    def _create_quality_gates_template(self) -> Dict[str, Any]:
        """Create quality gates configuration template."""
        return {
            "id": "quality-gates-setup",
            "title": "Quality Gates Configuration",
            "description": "Configure quality standards and validation rules",
            "questions": [
                {
                    "id": "constitution.test_coverage_target",
                    "text": "What is your target test coverage percentage?",
                    "type": "number_input",
                    "required": True,
                    "default_value": 85,
                    "validation": {
                        "rules": ["min_value", "max_value"],
                        "parameters": {"min_value": 0, "max_value": 100},
                    },
                },
                {
                    "id": "constitution.enforce_tdd",
                    "text": "Enforce Test-Driven Development (TDD)?",
                    "type": "boolean",
                    "required": True,
                    "default_value": True,
                },
                {
                    "id": "quality.auto_linting",
                    "text": "Enable automatic code linting?",
                    "type": "boolean",
                    "required": True,
                    "default_value": True,
                },
                {
                    "id": "quality.auto_formatting",
                    "text": "Enable automatic code formatting?",
                    "type": "boolean",
                    "required": True,
                    "default_value": True,
                },
            ],
        }

    def _create_skill_selection_template(self) -> Dict[str, Any]:
        """Create skill selection template."""
        return {
            "id": "skill-selection",
            "title": "Skill Selection",
            "description": "Select MoAI skills to enable for your project",
            "questions": [
                {
                    "id": "skills.enabled",
                    "text": "Which skills would you like to enable?",
                    "type": "multi_choice",
                    "required": True,
                    "options": [
                        {
                            "label": "Python Language Support",
                            "value": "moai-lang-python",
                            "description": "Python-specific patterns and best practices",
                        },
                        {
                            "label": "TypeScript Support",
                            "value": "moai-lang-typescript",
                            "description": "TypeScript and JavaScript development tools",
                        },
                        {
                            "label": "Backend Architecture",
                            "value": "moai-domain-backend",
                            "description": "API design and backend patterns",
                        },
                        {
                            "label": "Frontend Development",
                            "value": "moai-domain-frontend",
                            "description": "React, Vue, and frontend frameworks",
                        },
                        {
                            "label": "Security Analysis",
                            "value": "moai-security-owasp",
                            "description": "OWASP security validation and best practices",
                        },
                        {
                            "label": "Performance Engineering",
                            "value": "moai-performance-engineering",
                            "description": "Performance optimization and analysis",
                        },
                        {
                            "label": "Documentation Generation",
                            "value": "moai-workflow-docs",
                            "description": "Automated documentation generation",
                        },
                    ],
                }
            ],
        }

    def _ask_question(self, question: Question) -> Any:
        """
        Ask a single question to the user.
        This is a placeholder for actual AskUserQuestion integration.
        """
        # This would integrate with Claude Code's AskUserQuestion
        # For now, return a mock response
        if question.type == QuestionType.BOOLEAN:
            return question.default_value or False
        elif question.type == QuestionType.SINGLE_CHOICE and question.options:
            return question.options[0].value
        elif question.type == QuestionType.TEXT_INPUT:
            return "default_response"
        else:
            return None

    def _validate_response(self, value: Any, validation: ValidationConfig) -> "ValidationResult":
        """Validate a response value."""
        # Placeholder for validation logic
        return ValidationResult(is_valid=True, error=None)

    def _evaluate_condition(self, condition: Dict[str, Any], context: Dict[str, Any]) -> bool:
        """Evaluate conditional logic for questions."""
        # Simple implementation for now
        for key, expected in condition.items():
            if key not in context:
                return False

            actual = context[key]
            if isinstance(expected, list):
                if actual not in expected:
                    return False
            elif actual != expected:
                return False

        return True

    def _set_nested_value(self, config: Dict[str, Any], path: str, value: Any):
        """Set a nested value in the configuration dict."""
        keys = path.split(".")
        current = config

        for key in keys[:-1]:
            if key not in current:
                current[key] = {}
            current = current[key]

        current[keys[-1]] = value

    def _parse_conditional_questions(self, conditional_configs: List[Dict[str, Any]]) -> List[Question]:
        """Parse conditional questions from configuration."""
        questions = []
        for config in conditional_configs:
            question = Question(
                id=config["id"],
                text=config["text"],
                type=QuestionType(config["type"]),
                options=[QuestionOption(opt["label"], opt["value"]) for opt in config.get("options", [])],
            )
            questions.append(question)

        return questions

    def _save_template_to_file(self, template: Dict[str, Any], file_path: Path):
        """Save a template to file."""
        try:
            with open(file_path, "w", encoding="utf-8") as f:
                json.dump(template, f, indent=2, ensure_ascii=False)
        except Exception as e:
            logger.error(f"Error saving template to {file_path}: {e}")


@dataclass
class ValidationResult:
    """Result of response validation."""

    is_valid: bool
    error: Optional[str] = None
    warnings: List[str] = field(default_factory=list)


# Integration helper functions for MoAI skills


def create_project_initializer(config: Dict[str, Any]) -> BatchQuestionsManager:
    """Create a project initializer with standard templates."""
    return BatchQuestionsManager(config)


def ask_skill_questions(skill_name: str, context: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
    """
    Ask skill-specific questions and return responses.

    Args:
        skill_name: Name of the skill
        context: Existing context for conditional questions

    Returns:
        Dict[str, Any]: User responses
    """
    manager = BatchQuestionsManager()
    template = manager.get_skill_specific_template(skill_name)

    if not template:
        logger.warning(f"No template found for skill: {skill_name}")
        return {}

    return manager.execute_batch(template, context or {})
