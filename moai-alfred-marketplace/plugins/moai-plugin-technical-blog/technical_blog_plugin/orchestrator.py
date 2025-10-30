"""
Blog Write Orchestrator: Coordinates 7 specialist agents based on parsed directives
Handles CREATE, OPTIMIZE, and LIST modes
"""

from dataclasses import dataclass
from typing import List, Dict, Optional
from .parser import ParsedDirective, Mode


@dataclass
class AgentTask:
    """Task assigned to a specialist agent"""
    agent_id: str
    agent_name: str
    task_type: str  # "strategy", "write", "seo", "code", "visual", "format"
    can_parallelize: bool = True


class BlogWriteOrchestrator:
    """Orchestrates blog creation workflow using 7 specialist agents"""

    AGENTS = {
        "technical-content-strategist": {
            "name": "Technical Content Strategist",
            "model": "Sonnet",
            "type": "strategy"
        },
        "technical-writer": {
            "name": "Technical Writer",
            "model": "Haiku",
            "type": "write"
        },
        "seo-discoverability-specialist": {
            "name": "SEO & Discoverability Specialist",
            "model": "Haiku",
            "type": "seo"
        },
        "code-example-curator": {
            "name": "Code Example Curator",
            "model": "Haiku",
            "type": "code"
        },
        "visual-content-designer": {
            "name": "Visual Content Designer",
            "model": "Haiku",
            "type": "visual"
        },
        "markdown-formatter": {
            "name": "Markdown Formatter",
            "model": "Haiku",
            "type": "format"
        },
        "template-workflow-coordinator": {
            "name": "Template Workflow Coordinator",
            "model": "Sonnet",
            "type": "coordinator"
        }
    }

    @staticmethod
    def orchestrate(parsed_directive: ParsedDirective) -> Dict:
        """
        Orchestrate agent execution based on parsed directive

        Args:
            parsed_directive: Parsed directive from parser

        Returns:
            Orchestration plan with agent execution sequence
        """
        if parsed_directive.mode == Mode.LIST:
            return BlogWriteOrchestrator._handle_list_mode()

        elif parsed_directive.mode == Mode.OPTIMIZE:
            return BlogWriteOrchestrator._handle_optimize_mode(parsed_directive)

        elif parsed_directive.mode == Mode.CREATE:
            return BlogWriteOrchestrator._handle_create_mode(parsed_directive)

        return {"error": "Unknown mode"}

    @staticmethod
    def _handle_list_mode() -> Dict:
        """Handle LIST mode: Show available templates"""
        return {
            "mode": "list",
            "templates": [
                {
                    "id": "tutorial",
                    "name": "Tutorial Template",
                    "description": "Step-by-step learning guide",
                    "best_for": "Teaching concepts, hands-on learning"
                },
                {
                    "id": "case-study",
                    "name": "Case Study Template",
                    "description": "Problem → Solution → Results",
                    "best_for": "Sharing real-world success stories, metrics"
                },
                {
                    "id": "howto",
                    "name": "How-to Guide Template",
                    "description": "Task-oriented actionable steps",
                    "best_for": "Practical task guides, troubleshooting"
                },
                {
                    "id": "announcement",
                    "name": "Announcement Template",
                    "description": "Feature/product introduction",
                    "best_for": "Announcing new tools, features, projects"
                },
                {
                    "id": "comparison",
                    "name": "Comparison Template",
                    "description": "Tool/framework analysis",
                    "best_for": "Comparing options, decision guidance"
                }
            ]
        }

    @staticmethod
    def _handle_optimize_mode(parsed_directive: ParsedDirective) -> Dict:
        """Handle OPTIMIZE mode: Improve existing post"""
        agents = [
            AgentTask(
                agent_id="seo-discoverability-specialist",
                agent_name="SEO & Discoverability Specialist",
                task_type="seo",
                can_parallelize=True
            ),
            AgentTask(
                agent_id="markdown-formatter",
                agent_name="Markdown Formatter",
                task_type="format",
                can_parallelize=True
            )
        ]

        return {
            "mode": "optimize",
            "file_path": parsed_directive.file_path,
            "workflow": "parallel",
            "agents": [
                {
                    "agent_id": task.agent_id,
                    "agent_name": task.agent_name,
                    "task": f"Optimize {parsed_directive.file_path}"
                }
                for task in agents
            ],
            "steps": [
                "Load existing markdown file",
                "Run SEO optimization (meta tags, hashtags, llms.txt)",
                "Run markdown linting and auto-fixes",
                "Generate optimization report",
                "Output updated file"
            ]
        }

    @staticmethod
    def _handle_create_mode(parsed_directive: ParsedDirective) -> Dict:
        """Handle CREATE mode: Write new blog post"""
        # Step 1: Coordinator thinks about what to do
        coordinator_task = {
            "phase": 1,
            "agent": "template-workflow-coordinator",
            "task": f"Parse directive and plan workflow",
            "input": {
                "template": parsed_directive.template_id,
                "topic": parsed_directive.topic,
                "difficulty": parsed_directive.difficulty
            }
        }

        # Step 2: Content Strategist plans structure
        strategist_task = {
            "phase": 2,
            "agent": "technical-content-strategist",
            "task": "Analyze audience and define content structure",
            "dependencies": ["phase_1"],
            "parallel": False
        }

        # Step 3: 4 agents work in parallel
        parallel_agents = [
            {
                "phase": 3,
                "agent": "technical-writer",
                "task": "Write blog post content",
                "dependencies": ["phase_2"],
                "parallel": True
            },
            {
                "phase": 3,
                "agent": "code-example-curator",
                "task": "Generate code examples",
                "dependencies": ["phase_2"],
                "parallel": True
            },
            {
                "phase": 3,
                "agent": "seo-discoverability-specialist",
                "task": "Optimize for SEO and discoverability",
                "dependencies": ["phase_2"],
                "parallel": True
            },
            {
                "phase": 3,
                "agent": "visual-content-designer",
                "task": "Design images and diagrams",
                "dependencies": ["phase_2"],
                "parallel": True
            }
        ]

        # Step 4: Markdown Formatter validates
        formatter_task = {
            "phase": 4,
            "agent": "markdown-formatter",
            "task": "Validate and fix markdown",
            "dependencies": ["phase_3"],
            "parallel": False
        }

        # Step 5: Coordinator assembles final output
        assembly_task = {
            "phase": 5,
            "agent": "template-workflow-coordinator",
            "task": "Assemble and verify final output",
            "dependencies": ["phase_4"],
            "parallel": False
        }

        return {
            "mode": "create",
            "template_id": parsed_directive.template_id,
            "topic": parsed_directive.topic,
            "difficulty": parsed_directive.difficulty or "beginner",
            "workflow": [
                coordinator_task,
                strategist_task,
                *parallel_agents,
                formatter_task,
                assembly_task
            ],
            "execution_summary": {
                "total_phases": 5,
                "parallel_opportunities": 1,
                "sequential_tasks": 4,
                "estimated_duration": "3-5 minutes"
            }
        }

    @staticmethod
    def get_template(template_id: str) -> Optional[str]:
        """Get template file path"""
        templates = {
            "tutorial": "templates/tutorial.md",
            "case-study": "templates/case-study.md",
            "howto": "templates/howto.md",
            "announcement": "templates/announcement.md",
            "comparison": "templates/comparison.md"
        }
        return templates.get(template_id)


# Example usage
if __name__ == "__main__":
    from .parser import DirectiveParser

    directives = [
        "Next.js 15 초보자 튜토리얼 작성",
        "./posts/test.md 최적화",
        "템플릿 목록"
    ]

    parser = DirectiveParser()
    orchestrator = BlogWriteOrchestrator()

    for directive in directives:
        parsed = parser.parse(directive)
        plan = orchestrator.orchestrate(parsed)
        print(f"Directive: {directive}")
        print(f"Plan: {plan}")
        print()
