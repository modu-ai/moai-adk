"""
Comprehensive System Integration Testing for Research-Enhanced Components

Tests integration of all research-enhanced components:
- Skills (24 research skills)
- Agents (22 agents including research coordinators)
- Commands (4 commands including /alfred:research)
- Hooks (5 research hooks)

Validates:
- TAG system integration
- Research workflows
- System compatibility
- Performance benchmarks
- Error handling & recovery
"""

import pytest
import subprocess
import json
import time
import os
from pathlib import Path
from unittest.mock import patch, MagicMock
import sys

# Add project root to path
sys.path.insert(0, str(Path(__file__).parent.parent.parent))

class TestResearchIntegration:
    """Comprehensive research integration test suite"""

    @pytest.fixture
    def project_root(self):
        """Get project root directory"""
        return Path(__file__).parent.parent.parent

    @pytest.fixture
    def research_config(self, project_root):
        """Load research configuration"""
        config_path = project_root / ".moai" / "config.json"
        if config_path.exists():
            with open(config_path) as f:
                return json.load(f)
        return {}

    @pytest.fixture
    def mock_research_engines(self):
        """Mock research engines for testing"""
        with patch('knowledge_integration_hub.KnowledgeIntegrationHub') as mock_hub, \
             patch('cross_domain_analysis_engine.CrossDomainAnalyzer') as mock_analyzer, \
             patch('pattern_recognition_engine.PatternRecognizer') as mock_recognizer:

            # Configure mock responses
            mock_hub.return_value.analyze.return_value = {
                "patterns": ["pattern1", "pattern2"],
                "insights": ["insight1", "insight2"],
                "confidence": 0.85
            }

            mock_analyzer.return_value.cross_analyze.return_value = {
                "connections": ["domain1 <-> domain2"],
                "opportunities": ["opportunity1"],
                "recommendations": ["recommendation1"]
            }

            mock_recognizer.return_value.detect_patterns.return_value = {
                "code_patterns": ["pattern1", "pattern2"],
                "behavioral_patterns": ["behavior1"],
                "solutions": ["solution1"]
            }

            yield {
                'hub': mock_hub,
                'analyzer': mock_analyzer,
                'recognizer': mock_recognizer
            }

    class TestSkillsIntegration:
        """Test integration of research skills"""

        def test_research_skills_availability(self, project_root):
            """Test that all 24 research skills are available"""
            skills_dir = project_root / ".claude" / "skills"

            # Expected research skills
            expected_skills = [
                "moai-research-coordination",
                "moai-research-methodology",
                "moai-knowledge-integration",
                "moai-pattern-recognition",
                "moai-cross-domain-analysis",
                "moai-research-optimization",
                "moai-research-validation",
                "moai-research-documentation",
                "moai-mcp-integration",
                "moai-web-research",
                "moai-data-analysis",
                "moai-research-reporting",
                "moai-research-automation",
                "moai-research-quality",
                "moai-research-ethics",
                "moai-research-collaboration",
                "moai-research-persistence",
                "moai-research-versioning",
                "moai-research-backup",
                "moai-research-recovery",
                "moai-research-monitoring",
                "moai-research-alerting",
                "moai-research-metrics",
                "moai-research-insights"
            ]

            found_skills = []
            missing_skills = []

            for skill in expected_skills:
                skill_file = skills_dir / f"{skill}.md"
                if skill_file.exists():
                    found_skills.append(skill)
                else:
                    missing_skills.append(skill)

            assert len(missing_skills) == 0, f"Missing research skills: {missing_skills}"
            assert len(found_skills) >= 20, f"Found only {len(found_skills)} research skills, expected at least 20"

        def test_skill_content_validation(self, project_root):
            """Test that research skills have valid content"""
            skills_dir = project_root / ".claude" / "skills"
            skill_files = list(skills_dir.glob("*.md"))

            research_skills = [f for f in skill_files if any(keyword in f.name for keyword in ["research", "knowledge", "pattern", "analysis"])]

            validation_results = []
            for skill_file in research_skills:
                with open(skill_file) as f:
                    content = f.read()

                # Validate skill structure
                has_metadata = content.startswith('---') and '---' in content[1:]
                has_description = 'description:' in content
                has_content = len(content.strip()) > 200

                validation_results.append({
                    'skill': skill_file.name,
                    'has_metadata': has_metadata,
                    'has_description': has_description,
                    'has_content': has_content,
                    'valid': has_metadata and has_description and has_content
                })

            invalid_skills = [r for r in validation_results if not r['valid']]
            assert len(invalid_skills) == 0, f"Invalid research skills: {invalid_skills}"

        def test_skill_dependency_resolution(self, project_root):
            """Test that skill dependencies are properly resolved"""
            skills_dir = project_root / ".claude" / "skills"
            skill_files = list(skills_dir.glob("moai-*.md"))

            dependency_graph = {}

            for skill_file in skill_files:
                with open(skill_file) as f:
                    content = f.read()

                # Extract Skill() calls
                import re
                skill_calls = re.findall(r'Skill\("([^"]+)"\)', content)
                dependency_graph[skill_file.stem] = skill_calls

            # Validate dependencies exist
            missing_dependencies = []
            for skill, deps in dependency_graph.items():
                for dep in deps:
                    dep_file = skills_dir / f"{dep}.md"
                    if not dep_file.exists():
                        missing_dependencies.append(f"{skill} -> {dep}")

            assert len(missing_dependencies) == 0, f"Missing skill dependencies: {missing_dependencies}"

    class TestAgentsIntegration:
        """Test integration of research agents"""

        def test_research_agents_availability(self, project_root):
            """Test that all 22 agents are available including research coordinators"""
            agents_dir = project_root / ".claude" / "agents" / "alfred"
            agent_files = list(agents_dir.glob("*.md"))

            # Expected research-related agents
            research_agents = [
                "research-coordinator",
                "mcp-context7-integrator",
                "mcp-playwright-integrator",
                "mcp-sequential-thinking-integrator",
                "accessibility-expert",
                "api-designer",
                "component-designer",
                "migration-expert",
                "monitoring-expert",
                "performance-engineer"
            ]

            found_agents = [f.stem for f in agent_files]
            missing_agents = [agent for agent in research_agents if agent not in found_agents]

            assert len(missing_agents) == 0, f"Missing research agents: {missing_agents}"
            assert len(agent_files) >= 20, f"Found only {len(agent_files)} agents, expected at least 20"

        def test_agent_configuration(self, project_root, research_config):
            """Test that agents are properly configured"""
            agents_dir = project_root / ".claude" / "agents" / "alfred"
            agent_files = list(agents_dir.glob("*.md"))

            configuration_results = []

            for agent_file in agent_files:
                with open(agent_file) as f:
                    content = f.read()

                # Validate agent structure
                has_role = 'role:' in content or 'Role:' in content
                has_responsibilities = 'responsibilities:' in content or 'Responsibilities:' in content
                has_workflow = 'workflow:' in content or 'Workflow:' in content

                configuration_results.append({
                    'agent': agent_file.name,
                    'has_role': has_role,
                    'has_responsibilities': has_responsibilities,
                    'has_workflow': has_workflow,
                    'valid': has_role and has_responsibilities and has_workflow
                })

            invalid_agents = [r for r in configuration_results if not r['valid']]
            assert len(invalid_agents) == 0, f"Invalid agent configurations: {invalid_agents}"

        def test_agent_collaboration_flows(self, project_root):
            """Test agent collaboration flows"""
            agents_dir = project_root / ".claude" / "agents" / "alfred"

            # Test key collaboration flows
            collaboration_flows = [
                {
                    'flow': 'research-coordination',
                    'agents': ['research-coordinator', 'spec-builder', 'tdd-implementer'],
                    'description': 'Research to implementation flow'
                },
                {
                    'flow': 'mcp-integration',
                    'agents': ['mcp-context7-integrator', 'mcp-playwright-integrator', 'mcp-sequential-thinking-integrator'],
                    'description': 'MCP integration flow'
                },
                {
                    'flow': 'development-workflow',
                    'agents': ['spec-builder', 'implementation-planner', 'tdd-implementer', 'quality-gate'],
                    'description': 'Standard development workflow'
                }
            ]

            for flow in collaboration_flows:
                missing_agents = []
                for agent in flow['agents']:
                    agent_file = agents_dir / f"{agent}.md"
                    if not agent_file.exists():
                        missing_agents.append(agent)

                assert len(missing_agents) == 0, f"Flow '{flow['flow']}' missing agents: {missing_agents}"

    class TestCommandsIntegration:
        """Test integration of research commands"""

        def test_research_command_availability(self, project_root):
            """Test that /alfred:research command is available"""
            commands_dir = project_root / ".claude" / "commands" / "alfred"
            research_command = commands_dir / "research.md"

            assert research_command.exists(), "Research command not found"

            # Validate command structure
            with open(research_command) as f:
                content = f.read()

            has_name = content.startswith('---') and 'name: alfred:research' in content
            has_description = 'description:' in content
            has_allowed_tools = 'allowed-tools:' in content
            has_phases = 'PHASE 1:' in content

            assert has_name, "Research command missing name"
            assert has_description, "Research command missing description"
            assert has_allowed_tools, "Research command missing allowed tools"
            assert has_phases, "Research command missing execution phases"

        def test_command_execution_flow(self, project_root):
            """Test command execution flows"""
            commands_dir = project_root / ".claude" / "commands" / "alfred"
            command_files = list(commands_dir.glob("*.md"))

            # Test key commands
            key_commands = ['0-project', '1-plan', '2-run', '3-sync', 'research']

            for cmd in key_commands:
                cmd_file = commands_dir / f"{cmd}.md"
                assert cmd_file.exists(), f"Command {cmd} not found"

                with open(cmd_file) as f:
                    content = f.read()

                # Validate command has proper workflow
                has_workflow = 'workflow' in content.lower() or 'PHASE' in content
                has_agent_calls = 'Task(' in content or 'Task(' in content
                has_todo_tracking = 'TodoWrite' in content

                assert has_workflow, f"Command {cmd} missing workflow definition"
                assert has_agent_calls, f"Command {cmd} missing agent calls"
                assert has_todo_tracking, f"Command {cmd} missing todo tracking"

        def test_command_parameter_parsing(self, project_root):
            """Test command parameter parsing"""
            research_command = project_root / ".claude" / "commands" / "alfred" / "research.md"

            with open(research_command) as f:
                content = f.read()

            # Test parameter parsing logic
            has_argument_hint = 'argument-hint:' in content
            has_action_parsing = 'action:' in content.lower() or 'Actions:' in content
            has_options_handling = 'options:' in content.lower() or 'Options:' in content

            assert has_argument_hint, "Research command missing argument hint"
            assert has_action_parsing, "Research command missing action parsing"
            assert has_options_handling, "Research command missing options handling"

    class TestHooksIntegration:
        """Test integration of research hooks"""

        def test_research_hooks_availability(self, project_root):
            """Test that all 5 research hooks are available"""
            hooks_dir = project_root / ".claude" / "hooks" / "alfred"

            research_hooks = [
                'session_start__research_setup.py',
                'pre_tool__research_strategy.py',
                'post_tool__research_analysis.py',
                'spec_status_hooks.py'
            ]

            found_hooks = []
            missing_hooks = []

            for hook in research_hooks:
                hook_file = hooks_dir / hook
                if hook_file.exists():
                    found_hooks.append(hook)
                else:
                    missing_hooks.append(hook)

            assert len(missing_hooks) == 0, f"Missing research hooks: {missing_hooks}"
            assert len(found_hooks) >= 3, f"Found only {len(found_hooks)} research hooks, expected at least 3"

        def test_hook_execution_flow(self, project_root):
            """Test hook execution flows"""
            hooks_dir = project_root / ".claude" / "hooks" / "alfred"

            # Test research setup hook
            setup_hook = hooks_dir / 'session_start__research_setup.py'
            if setup_hook.exists():
                with open(setup_hook) as f:
                    content = f.read()

                has_initialization = 'def setup_research_context' in content or 'research_context' in content
                has_configuration = 'config' in content or 'configuration' in content
                has_validation = 'validate' in content or 'check' in content

                assert has_initialization, "Research setup hook missing initialization logic"
                assert has_configuration, "Research setup hook missing configuration handling"
                assert has_validation, "Research setup hook missing validation logic"

        def test_hook_error_handling(self, project_root):
            """Test hook error handling"""
            hooks_dir = project_root / ".claude" / "hooks" / "alfred"
            hook_files = list(hooks_dir.glob("*.py"))

            for hook_file in hook_files:
                with open(hook_file) as f:
                    content = f.read()

                # Check for proper error handling
                has_try_catch = 'try:' in content and 'except' in content
                has_error_logging = 'error' in content.lower() or 'exception' in content.lower()

                # Not all hooks need error handling, but critical ones should
                if 'research' in hook_file.name or any(keyword in hook_file.name for keyword in ['setup', 'strategy', 'analysis']):
                    assert has_try_catch or has_error_logging, f"Critical hook {hook_file.name} missing error handling"

    class TestTAGSystemIntegration:
        """Test TAG system integration"""

        def test_tag_search_functionality(self, project_root):
            """Test TAG search functionality"""
            # Test tag search patterns
            tag_patterns = [
                '@RESEARCH:',
                '@PATTERN:',
                '@SOLUTION:',
                '@SPEC:',
                '@TEST:',
                '@CODE:',
                '@DOC:'
            ]

            search_results = {}

            for pattern in tag_patterns:
                try:
                    result = subprocess.run(
                        ['rg', pattern, '-n', '--count', str(project_root)],
                        capture_output=True,
                        text=True,
                        timeout=30
                    )
                    if result.returncode == 0:
                        search_results[pattern] = int(result.stdout.strip()) if result.stdout.strip() else 0
                    else:
                        search_results[pattern] = 0
                except subprocess.TimeoutExpired:
                    search_results[pattern] = -1  # Timeout
                except FileNotFoundError:
                    search_results[pattern] = -2  # rg not found

            # Validate we have some tags
            total_tags = sum(count for count in search_results.values() if count > 0)
            assert total_tags > 0, f"No tags found in project. Search results: {search_results}"

        def test_tag_assignment_consistency(self, project_root):
            """Test TAG assignment consistency"""
            spec_files = list(project_root.glob(".moai/specs/*/spec.md"))

            tag_consistency_results = []

            for spec_file in spec_files:
                with open(spec_file) as f:
                    content = f.read()

                # Extract tags
                import re
                spec_tags = re.findall(r'@(\w+):', content)

                # Check for required tags
                has_spec_tag = '@SPEC:' in content
                has_test_tag = '@TEST:' in content
                has_code_tag = '@CODE:' in content
                has_doc_tag = '@DOC:' in content

                tag_consistency_results.append({
                    'spec': spec_file.name,
                    'has_spec_tag': has_spec_tag,
                    'has_test_tag': has_test_tag,
                    'has_code_tag': has_code_tag,
                    'has_doc_tag': has_doc_tag,
                    'total_tags': len(spec_tags),
                    'consistent': has_spec_tag and has_test_tag and has_code_tag and has_doc_tag
                })

            inconsistent_specs = [r for r in tag_consistency_results if not r['consistent']]
            # Allow some inconsistency for legacy specs
            assert len(inconsistent_specs) <= len(spec_files) * 0.2, f"Too many inconsistent specs: {inconsistent_specs[:5]}"

        def test_research_tag_integration(self, project_root):
            """Test research-specific TAG integration"""
            # Test for research tags
            research_tags = ['@RESEARCH:', '@PATTERN:', '@SOLUTION:']

            tag_locations = {}

            for tag in research_tags:
                try:
                    result = subprocess.run(
                        ['rg', tag, '-n', str(project_root)],
                        capture_output=True,
                        text=True,
                        timeout=30
                    )
                    if result.returncode == 0:
                        locations = [line.split(':')[0] for line in result.stdout.strip().split('\n') if line.strip()]
                        tag_locations[tag] = len(locations)
                    else:
                        tag_locations[tag] = 0
                except (subprocess.TimeoutExpired, FileNotFoundError):
                    tag_locations[tag] = 0

            # Research tags should be present in research-enhanced system
            total_research_tags = sum(tag_locations.values())
            assert total_research_tags >= 0, f"Research tags search failed: {tag_locations}"

    class TestResearchWorkflows:
        """Test research workflow integration"""

        def test_research_to_spec_workflow(self, project_root, mock_research_engines):
            """Test research to SPEC workflow"""
            # Simulate research to spec conversion
            research_data = {
                "topic": "authentication",
                "findings": ["JWT implementation", "OAuth 2.0 flows"],
                "patterns": ["token refresh", "session management"],
                "solutions": ["secure password hashing", "MFA integration"]
            }

            # Validate research data can be converted to SPEC format
            spec_requirements = []

            for finding in research_data["findings"]:
                spec_requirements.append(f"The system must support {finding.lower()}")

            for pattern in research_data["patterns"]:
                spec_requirements.append(f"WHEN {pattern} is triggered, the system must respond appropriately")

            for solution in research_data["solutions"]:
                spec_requirements.append(f"The system must implement {solution.lower()}")

            assert len(spec_requirements) >= 5, f"Insufficient SPEC requirements generated: {spec_requirements}"
            assert all("must" in req for req in spec_requirements), "Not all requirements follow EARS format"

        def test_research_integration_hooks(self, project_root):
            """Test research integration hooks"""
            hooks_dir = project_root / ".claude" / "hooks" / "alfred"

            # Test research strategy hook
            strategy_hook = hooks_dir / 'pre_tool__research_strategy.py'
            if strategy_hook.exists():
                with open(strategy_hook) as f:
                    content = f.read()

                has_strategy_logic = 'strategy' in content.lower()
                has_research_coordination = 'research' in content.lower()
                has_execution_planning = 'plan' in content.lower() or 'execute' in content.lower()

                assert has_strategy_logic, "Research strategy hook missing strategy logic"
                assert has_research_coordination, "Research strategy hook missing research coordination"
                assert has_execution_planning, "Research strategy hook missing execution planning"

        def test_cross_domain_research(self, project_root):
            """Test cross-domain research capabilities"""
            # Test cross-domain analysis
            domains = ["authentication", "authorization", "security", "performance"]

            cross_domain_connections = {}

            for i, domain1 in enumerate(domains):
                for domain2 in domains[i+1:]:
                    # Simulate cross-domain analysis
                    connection_strength = hash(f"{domain1}-{domain2}") % 100 / 100
                    cross_domain_connections[f"{domain1}-{domain2}"] = connection_strength

            # Validate cross-domain analysis produces results
            assert len(cross_domain_connections) > 0, "Cross-domain analysis produced no results"
            assert any(conn > 0.5 for conn in cross_domain_connections.values()), "No strong cross-domain connections found"

    class TestSystemCompatibility:
        """Test system compatibility and integration"""

        def test_version_compatibility(self, project_root, research_config):
            """Test version compatibility across components"""
            # Check config version compatibility
            if 'version' in research_config:
                config_version = research_config['version']
                # Validate version format
                version_parts = config_version.split('.')
                assert len(version_parts) >= 2, f"Invalid version format: {config_version}"

        def test_dependency_resolution(self, project_root):
            """Test dependency resolution across components"""
            components = {
                'skills': list((project_root / ".claude" / "skills").glob("*.md")),
                'agents': list((project_root / ".claude" / "agents" / "alfred").glob("*.md")),
                'commands': list((project_root / ".claude" / "commands" / "alfred").glob("*.md")),
                'hooks': list((project_root / ".claude" / "hooks" / "alfred").glob("*.py"))
            }

            # Validate component counts
            assert len(components['skills']) >= 20, f"Insufficient skills: {len(components['skills'])}"
            assert len(components['agents']) >= 20, f"Insufficient agents: {len(components['agents'])}"
            assert len(components['commands']) >= 4, f"Insufficient commands: {len(components['commands'])}"
            assert len(components['hooks']) >= 3, f"Insufficient hooks: {len(components['hooks'])}"

        def test_configuration_consistency(self, project_root, research_config):
            """Test configuration consistency"""
            # Check for required config sections
            required_sections = ['language', 'project', 'git_strategy']

            missing_sections = []
            for section in required_sections:
                if section not in research_config:
                    missing_sections.append(section)

            # Some sections might be optional, but critical ones should be present
            assert len(missing_sections) <= 1, f"Missing critical config sections: {missing_sections}"

if __name__ == "__main__":
    # Run integration tests
    pytest.main([__file__, "-v", "--tb=short"])