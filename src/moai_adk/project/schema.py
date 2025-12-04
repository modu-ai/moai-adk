"""
Tab schema v3.0.0 definition and loading

Defines the 3-tab configuration schema with:
- Tab 1: Quick Start (10 questions in 4 batches)
- Tab 2: Documentation (1-2 questions, conditional)
- Tab 3: Git Automation (0-4 questions, conditional based on mode)
"""

from typing import Any, Dict


def load_tab_schema() -> Dict[str, Any]:
    """Load and return the tab schema v3.0.0"""
    return {
        "version": "3.0.0",
        "tabs": [
            _create_tab1_quick_start(),
            _create_tab2_documentation(),
            _create_tab3_git_automation(),
        ],
    }


def _create_tab1_quick_start() -> Dict[str, Any]:
    """
    Tab 1: Quick Start
    4 batches, 10 questions total
    """
    return {
        "id": "tab_1_quick_start",
        "label": "Essential Setup",
        "description": "Configure project basics",
        "batches": [
            # Batch 1.1: Identity & Language (3 questions)
            {
                "id": "batch_1_1_identity",
                "header": "Identity",
                "batch_number": 1,
                "total_batches": 4,
                "questions": [
                    {
                        "id": "user_name",
                        "question": "What name should Alfred use?",
                        "type": "text_input",
                        "required": True,
                        "smart_default": "User",
                        "options": [
                            {"label": "Enter Name", "value": "custom"},
                            {"label": "Use Default", "value": "default"},
                        ],
                    },
                    {
                        "id": "conversation_language",
                        "question": "Conversation language?",
                        "type": "select_single",
                        "required": True,
                        "options": [
                            {
                                "label": "Korean (ko)",
                                "value": "ko",
                                "description": "Korean language",
                            },
                            {
                                "label": "English (en)",
                                "value": "en",
                                "description": "English language",
                            },
                            {
                                "label": "Japanese (ja)",
                                "value": "ja",
                                "description": "Japanese language",
                            },
                            {
                                "label": "Chinese (zh)",
                                "value": "zh",
                                "description": "Chinese language",
                            },
                        ],
                    },
                    {
                        "id": "agent_prompt_language",
                        "question": "Agent prompt language?",
                        "type": "select_single",
                        "required": True,
                        "smart_default": "en",
                        "options": [
                            {
                                "label": "English",
                                "value": "en",
                                "description": "Recommended for best results",
                            },
                            {
                                "label": "Korean",
                                "value": "ko",
                                "description": "Korean language",
                            },
                        ],
                    },
                ],
            },
            # Batch 1.2: Project Basics (3 questions)
            {
                "id": "batch_1_2_project_basics",
                "header": "Project",
                "batch_number": 2,
                "total_batches": 4,
                "questions": [
                    {
                        "id": "project_name",
                        "question": "Project Name?",
                        "type": "text_input",
                        "required": True,
                        "smart_default": "my-project",
                        "options": [
                            {"label": "Enter Name", "value": "custom"},
                            {"label": "Use Default", "value": "default"},
                        ],
                    },
                    {
                        "id": "github_profile_name",
                        "question": "GitHub Profile Name? (e.g., @GoosLab)",
                        "type": "text_input",
                        "required": False,
                        "smart_default": "",
                        "options": [
                            {"label": "Enter GitHub Profile", "value": "custom"},
                            {"label": "Skip (set later)", "value": "skip"},
                        ],
                    },
                    {
                        "id": "project_description",
                        "question": "Project Description?",
                        "type": "text_input",
                        "required": False,
                        "smart_default": "",
                        "options": [
                            {"label": "Enter Description", "value": "custom"},
                            {"label": "Skip", "value": "skip"},
                        ],
                    },
                ],
            },
            # Batch 1.3: Development Mode (2 questions)
            {
                "id": "batch_1_3_development_mode",
                "header": "Development",
                "batch_number": 3,
                "total_batches": 4,
                "questions": [
                    {
                        "id": "git_strategy_mode",
                        "question": "Git workflow mode?",
                        "type": "select_single",
                        "required": True,
                        "options": [
                            {
                                "label": "Personal",
                                "value": "personal",
                                "description": "Solo development",
                            },
                            {
                                "label": "Team",
                                "value": "team",
                                "description": "Team collaboration",
                            },
                            {
                                "label": "Hybrid",
                                "value": "hybrid",
                                "description": "Solo + Team flexible",
                            },
                        ],
                    },
                    {
                        "id": "git_strategy_workflow",
                        "question": "Branching workflow type?",
                        "type": "select_single",
                        "required": True,
                        "conditional_mapping": {
                            "personal": ["github-flow", "trunk-based"],
                            "team": ["git-flow", "github-flow"],
                            "hybrid": ["github-flow", "git-flow"],
                        },
                        "smart_default_mapping": {
                            "personal": "github-flow",
                            "team": "git-flow",
                            "hybrid": "github-flow",
                        },
                        "options": [
                            {
                                "label": "GitHub Flow",
                                "value": "github-flow",
                                "description": "Simple flow",
                            },
                            {
                                "label": "Git Flow",
                                "value": "git-flow",
                                "description": "Complex flow",
                            },
                            {
                                "label": "Trunk-Based",
                                "value": "trunk-based",
                                "description": "Continuous delivery",
                            },
                        ],
                    },
                ],
            },
            # Batch 1.4: Quality Standards (2 questions)
            {
                "id": "batch_1_4_quality_standards",
                "header": "Quality",
                "batch_number": 4,
                "total_batches": 4,
                "questions": [
                    {
                        "id": "test_coverage_target",
                        "question": "Test coverage target?",
                        "type": "number_input",
                        "required": True,
                        "smart_default": 90,
                        "min": 0,
                        "max": 100,
                        "options": [],
                    },
                    {
                        "id": "enforce_tdd",
                        "question": "Enforce TDD?",
                        "type": "select_single",
                        "required": True,
                        "smart_default": True,
                        "options": [
                            {
                                "label": "Yes",
                                "value": True,
                                "description": "Strict TDD",
                            },
                            {
                                "label": "No",
                                "value": False,
                                "description": "Optional TDD",
                            },
                        ],
                    },
                ],
            },
        ],
    }


def _create_tab2_documentation() -> Dict[str, Any]:
    """
    Tab 2: Documentation
    2 batches, 1 required + 1 conditional question
    """
    return {
        "id": "tab_2_documentation",
        "label": "Documentation",
        "description": "Project document generation",
        "batches": [
            # Batch 2.1: Documentation Choice (1 question)
            {
                "id": "batch_2_1_documentation_choice",
                "header": "Docs",
                "batch_number": 1,
                "total_batches": 2,
                "questions": [
                    {
                        "id": "documentation_mode",
                        "question": "Documentation strategy?",
                        "type": "select_single",
                        "required": True,
                        "options": [
                            {
                                "label": "Quick Start - Skip for Now",
                                "value": "skip",
                                "description": "30 seconds - No docs now",
                            },
                            {
                                "label": "Full Documentation - Now",
                                "value": "full_now",
                                "description": "5-30min - Generate docs now",
                            },
                            {
                                "label": "Minimal - Auto-generate",
                                "value": "minimal",
                                "description": "1 min - Auto templates",
                            },
                        ],
                    },
                ],
            },
            # Batch 2.2: Documentation Depth (1 conditional question)
            {
                "id": "batch_2_2_documentation_depth",
                "header": "Depth",
                "batch_number": 2,
                "total_batches": 2,
                "show_if": "documentation_mode == 'full_now'",
                "questions": [
                    {
                        "id": "documentation_depth",
                        "question": "Brainstorming depth?",
                        "type": "select_single",
                        "required": True,
                        "show_if": "documentation_mode == 'full_now'",
                        "options": [
                            {
                                "label": "Quick",
                                "value": "quick",
                                "description": "5-10 min brainstorming",
                            },
                            {
                                "label": "Standard",
                                "value": "standard",
                                "description": "10-15 min brainstorming",
                            },
                            {
                                "label": "Deep",
                                "value": "deep",
                                "description": "25-30 min brainstorming",
                            },
                        ],
                    },
                ],
            },
        ],
    }


def _create_tab3_git_automation() -> Dict[str, Any]:
    """
    Tab 3: Git Automation
    2 conditional batches based on git_strategy.mode
    """
    return {
        "id": "tab_3_git_automation",
        "label": "Git",
        "description": "Git automation settings",
        "batches": [
            # Batch 3.1: Personal Git Settings (conditional)
            {
                "id": "batch_3_1_personal",
                "header": "Personal",
                "batch_number": 1,
                "total_batches": 2,
                "show_if": "git_strategy_mode == 'personal' OR git_strategy_mode == 'hybrid'",
                "questions": [
                    {
                        "id": "git_personal_auto_checkpoint",
                        "question": "Auto checkpoint?",
                        "type": "select_single",
                        "required": True,
                        "show_if": "git_strategy_mode == 'personal' OR git_strategy_mode == 'hybrid'",
                        "smart_default": "disabled",
                        "options": [
                            {
                                "label": "Disabled",
                                "value": "disabled",
                                "description": "No auto checkpoints",
                            },
                            {
                                "label": "Event-Driven",
                                "value": "event-driven",
                                "description": "Auto save on events",
                            },
                            {
                                "label": "Manual",
                                "value": "manual",
                                "description": "Manual saves only",
                            },
                        ],
                    },
                    {
                        "id": "git_personal_push_remote",
                        "question": "Push to remote?",
                        "type": "select_single",
                        "required": True,
                        "show_if": "git_strategy_mode == 'personal' OR git_strategy_mode == 'hybrid'",
                        "smart_default": False,
                        "options": [
                            {
                                "label": "Yes",
                                "value": True,
                                "description": "Auto push",
                            },
                            {
                                "label": "No",
                                "value": False,
                                "description": "Manual push",
                            },
                        ],
                    },
                ],
            },
            # Batch 3.2: Team Git Settings (conditional)
            {
                "id": "batch_3_1_team",
                "header": "Team",
                "batch_number": 2,
                "total_batches": 2,
                "show_if": "git_strategy_mode == 'team'",
                "questions": [
                    {
                        "id": "git_team_auto_pr",
                        "question": "Auto PR?",
                        "type": "select_single",
                        "required": True,
                        "show_if": "git_strategy_mode == 'team'",
                        "smart_default": False,
                        "options": [
                            {
                                "label": "Yes",
                                "value": True,
                                "description": "Auto create PR",
                            },
                            {
                                "label": "No",
                                "value": False,
                                "description": "Manual PR",
                            },
                        ],
                    },
                    {
                        "id": "git_team_draft_pr",
                        "question": "Draft PR?",
                        "type": "select_single",
                        "required": True,
                        "show_if": "git_strategy_mode == 'team'",
                        "smart_default": False,
                        "options": [
                            {
                                "label": "Yes",
                                "value": True,
                                "description": "Draft PR mode",
                            },
                            {
                                "label": "No",
                                "value": False,
                                "description": "Ready for review",
                            },
                        ],
                    },
                ],
            },
        ],
    }
