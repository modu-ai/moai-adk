"""
AskUserQuestion Integration Module

Provides seamless integration between the BatchQuestionsManager and Claude Code's
AskUserQuestion functionality, ensuring consistent user interaction patterns
across all MoAI skills.

Author: MoAI-ADK
Version: 1.0.0
"""

import logging
from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional, Union

from .batch_questions import BatchQuestionsManager, Question, QuestionType, UserResponse

logger = logging.getLogger(__name__)


@dataclass
class AskUserConfig:
    """Configuration for AskUserQuestion integration."""

    max_options_per_question: int = 10
    enable_descriptions: bool = True
    enable_conditional: bool = True
    validation_error_retries: int = 3
    timeout_seconds: Optional[int] = None
    default_multi_select: bool = False


class AskUserQuestionIntegrator:
    """
    Integrates BatchQuestionsManager with Claude Code's AskUserQuestion.

    This class provides a bridge between our question system and Claude Code's
    built-in user interaction capabilities, ensuring consistent UX across
    all MoAI skills.
    """

    def __init__(self, config: Optional[AskUserConfig] = None):
        """Initialize the integrator with optional configuration."""
        self.config = config or AskUserConfig()
        self._setup_ask_user_function()

    def _setup_ask_user_function(self):
        """Setup the AskUserQuestion function reference."""
        # This will be available when running in Claude Code environment
        # For now, we'll create a mock implementation
        try:
            # In actual Claude Code environment, this would be available
            from claude_code import ask_user_question as AskUserQuestion  # noqa: N812

            self.AskUserQuestion = AskUserQuestion
        except ImportError:
            # Mock implementation for development/testing
            self.AskUserQuestion = self._mock_ask_user_question
            logger.warning("Using mock AskUserQuestion - not in Claude Code environment")

    def _mock_ask_user_question(
        self, questions: List[Dict[str, Any]], answers: Optional[Dict[str, str]] = None
    ) -> Dict[str, str]:
        """Mock implementation of AskUserQuestion for development."""
        logger.info(f"Mock AskUserQuestion called with {len(questions)} questions")

        # Return mock answers
        mock_answers = {}
        for q in questions:
            if q.get("multiSelect", False):
                # For multi-select, return first option
                options = q.get("options", [])
                if options:
                    mock_answers[q["question"]] = [options[0]["label"]]
            else:
                # For single select, return first option
                options = q.get("options", [])
                if options:
                    mock_answers[q["question"]] = options[0]["label"]
                else:
                    mock_answers[q["question"]] = "Default answer"

        return mock_answers

    def convert_question_to_ask_user_format(self, question: Question) -> Dict[str, Any]:
        """
        Convert a Question object to AskUserQuestion format.

        Args:
            question: Question to convert

        Returns:
            Dict: AskUserQuestion compatible format
        """
        ask_user_q = {
            "question": question.text,
            "multiSelect": question.type == QuestionType.MULTI_CHOICE,
        }

        # Add header if available
        if question.metadata.get("header"):
            ask_user_q["header"] = question.metadata["header"]

        # Convert options
        if question.options:
            options = []
            for opt in question.options:
                option_dict = {
                    "label": opt.label,
                    "description": (opt.description if self.config.enable_descriptions else None),
                }

                # Only include non-None descriptions
                if option_dict["description"] is None:
                    del option_dict["description"]

                options.append(option_dict)

            ask_user_q["options"] = options

        # Handle text input questions (convert to choice with text input)
        elif question.type in [QuestionType.TEXT_INPUT, QuestionType.NUMBER_INPUT]:
            ask_user_q["options"] = [
                {
                    "label": "Enter value...",
                    "description": f"Type your {question.type.value}",
                }
            ]

        # Handle boolean questions
        elif question.type == QuestionType.BOOLEAN:
            ask_user_q["options"] = [
                {"label": "Yes", "description": "Confirm or accept"},
                {"label": "No", "description": "Decline or reject"},
            ]

        # Add validation metadata
        if question.validation and question.validation.required:
            ask_user_q["required"] = True

        return ask_user_q

    def process_ask_user_response(self, question: Question, response_value: Union[str, List[str]]) -> Any:
        """
        Process AskUserQuestion response back to our expected format.

        Args:
            question: Original question
            response_value: Response from AskUserQuestion

        Returns:
            Processed response value in our expected format
        """
        # Handle multi-select responses
        if isinstance(response_value, list) and question.type == QuestionType.MULTI_CHOICE:
            # Convert labels back to values
            selected_values = []
            for label in response_value:
                for option in question.options:
                    if option.label == label:
                        selected_values.append(option.value)
                        break
            return selected_values

        # Handle single-select responses
        elif isinstance(response_value, str):
            # For choice questions, convert label back to value
            if question.options:
                for option in question.options:
                    if option.label == response_value:
                        return option.value

            # For text input questions, return as-is
            elif question.type in [QuestionType.TEXT_INPUT, QuestionType.NUMBER_INPUT]:
                # Convert number input to appropriate type
                if question.type == QuestionType.NUMBER_INPUT:
                    try:
                        return int(response_value)
                    except ValueError:
                        try:
                            return float(response_value)
                        except ValueError:
                            return response_value  # Keep as string if conversion fails
                return response_value

            # For boolean questions
            elif question.type == QuestionType.BOOLEAN:
                return response_value.lower() in ["yes", "y", "true", "1"]

        return response_value

    def ask_question(self, question: Question, context: Optional[Dict[str, Any]] = None) -> UserResponse:
        """
        Ask a single question using AskUserQuestion.

        Args:
            question: Question to ask
            context: Current context for conditional questions

        Returns:
            UserResponse: The user's response
        """
        # Check if question should be asked based on conditional logic
        if question.conditional_on and context:
            should_ask = self._evaluate_condition(question.conditional_on, context)
            if not should_ask:
                # Return default response or skip
                return UserResponse(
                    question_id=question.id,
                    value=question.default_value,
                    metadata={"skipped": True, "reason": "conditional_not_met"},
                )

        # Convert to AskUserQuestion format
        ask_user_q = self.convert_question_to_ask_user_format(question)

        # Ask the question
        try:
            raw_response = self.AskUserQuestion([ask_user_q])
            response_value = raw_response.get(ask_user_q["question"])

            # Validate response
            if question.validation:
                validation_result = self._validate_response(response_value, question.validation)
                if not validation_result.is_valid:
                    raise ValueError(f"Invalid response: {validation_result.error}")

            # Process response
            processed_value = self.process_ask_user_response(question, response_value)

            return UserResponse(
                question_id=question.id,
                value=processed_value,
                metadata={"source": "ask_user_question", "raw_response": raw_response},
            )

        except Exception as e:
            logger.error(f"Error asking question {question.id}: {e}")
            # Return default value on error
            return UserResponse(
                question_id=question.id,
                value=question.default_value,
                metadata={"error": str(e), "fallback": True},
            )

    def ask_question_batch(
        self, questions: List[Question], context: Optional[Dict[str, Any]] = None
    ) -> Dict[str, UserResponse]:
        """
        Ask multiple questions in sequence using AskUserQuestion.

        Args:
            questions: List of questions to ask
            context: Current context for conditional questions

        Returns:
            Dict[str, UserResponse]: All user responses
        """
        if context is None:
            context = {}

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
            response = self.ask_question(question, context)
            responses[question.id] = response

            # Update context for subsequent questions
            if not response.metadata.get("skipped"):
                context[question.id] = response.value

        return responses

    def _evaluate_condition(self, condition: Dict[str, Any], context: Dict[str, Any]) -> bool:
        """Evaluate conditional logic for questions."""
        if not self.config.enable_conditional:
            return True

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

    def _validate_response(self, value: Any, validation) -> "ValidationResult":
        """Validate a response value."""
        # This would integrate with the validation system
        # For now, return a simple validation result
        return ValidationResult(is_valid=True, error=None)


@dataclass
class ValidationResult:
    """Result of response validation."""

    is_valid: bool
    error: Optional[str] = None
    warnings: List[str] = field(default_factory=list)


# Monkey patch BatchQuestionsManager to use AskUserQuestion integration


def enhance_batch_questions_manager():
    """Enhance BatchQuestionsManager with AskUserQuestion integration."""

    def execute_batch_with_ask_user(self, batch, context=None):
        """Execute batch using AskUserQuestion integration."""
        if context is None:
            context = {}

        integrator = AskUserQuestionIntegrator()
        questions = self.filter_questions(batch, context)
        return integrator.ask_question_batch(questions, context)

    # Patch the method
    BatchQuestionsManager.execute_batch = execute_batch_with_ask_user

    logger.info("Enhanced BatchQuestionsManager with AskUserQuestion integration")


# Auto-enhance when module is imported
enhance_batch_questions_manager()
