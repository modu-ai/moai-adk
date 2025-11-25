---
name: moai-core-agent-factory
description: Enterprise agent factory with intelligent expertise detection, automated agent creation, and dynamic orchestration capabilities
allowed-tools: Read, Write, Edit, Bash, Grep, Glob, TodoWrite, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
---

# Enterprise Agent Factory 🤖

Advanced agent creation and orchestration system that automates the generation, configuration, and deployment of specialized AI agents with intelligent expertise detection. Provides enterprise-grade agent management with dynamic skill allocation, performance monitoring, and continuous optimization.

## Quick Reference (30 seconds)

**Core Capabilities**:
- **Intelligent Agent Generation**: Automated agent creation with expertise-matched capabilities
- **Dynamic Expertise Detection**: AI-powered identification of optimal agent specializations
- **Enterprise Orchestration**: Multi-agent coordination with intelligent load balancing
- **Performance Optimization**: Real-time agent performance monitoring and auto-tuning
- **Scalable Architecture**: Support for hundreds of concurrent agents with resource management

**Key Patterns**:
1. **Expertise Mapping Engine** → Intelligent agent-task matching algorithms
2. **Dynamic Agent Generation** → Context-aware agent creation with specialized skills
3. **Multi-Agent Orchestration** → Intelligent coordination and resource allocation
4. **Performance Analytics** → Continuous optimization and learning from execution patterns

**When to Use**:
- Creating specialized agents for complex domain-specific tasks
- Scaling agent deployment across multiple projects or teams
- Optimizing agent performance and resource utilization
- Implementing intelligent agent coordination and workflow automation

## Implementation Guide

### Getting Started

**Basic Agent Creation**:
```python
# Initialize agent factory
agent_factory = AgentFactory(
    context7_integration=True,
    performance_monitoring=True,
    auto_optimization=True
)

# Detect required expertise for task
task_description = "Analyze financial data and generate investment recommendations"
expertise_profile = await agent_factory.detect_expertise(task_description)

# Generate specialized agent
financial_analyst = await agent_factory.create_agent(
    expertise=expertise_profile,
    capabilities=["data_analysis", "financial_modeling", "report_generation"],
    tools=["pandas", "numpy", "matplotlib", "yfinance"]
)

# Deploy and use agent
result = await financial_analyst.execute_task(task_data)
```

**Expertise Detection and Matching**:
```python
# Initialize expertise detection engine
expertise_engine = ExpertiseDetectionEngine(context7_client=context7_client)

# Analyze task complexity and requirements
task_analysis = await expertise_engine.analyze_task_requirements(
    task_description="Build a real-time chat application with WebSocket support",
    context={"tech_stack": ["python", "fastapi", "websocket"], "team_size": 3}
)

# Detect optimal expertise profile
expertise_profile = await expertise_engine.detect_optimal_expertise(
    task_analysis=task_analysis,
    existing_agents=agent_registry.get_active_agents()
)

print(f"Recommended expertise: {expertise_profile.primary_domain}")
print(f"Secondary domains: {expertise_profile.secondary_domains}")
print(f"Required tools: {expertise_profile.required_tools}")
print(f"Complexity level: {expertise_profile.complexity_level}")
```

### Core Components

#### 1. Agent Factory Engine

```python
class AgentFactory:
    """Enterprise-grade agent creation and management system"""

    def __init__(self, config: AgentFactoryConfig):
        self.expertise_engine = ExpertiseDetectionEngine()
        self.agent_generator = AgentGenerator()
        self.orchestrator = MultiAgentOrchestrator()
        self.performance_monitor = AgentPerformanceMonitor()
        self.context7_client = Context7Client()
        self.agent_registry = AgentRegistry()

    async def create_agent(
        self,
        expertise: ExpertiseProfile,
        capabilities: List[str],
        tools: List[str],
        configuration: Optional[AgentConfig] = None
    ) -> Agent:
        """Create a specialized agent based on expertise profile"""

        # Phase 1: Validate expertise requirements
        validation_result = await self._validate_expertise_requirements(expertise)
        if not validation_result.is_valid:
            raise AgentCreationError(f"Invalid expertise profile: {validation_result.errors}")

        # Phase 2: Generate agent configuration
        agent_config = await self._generate_agent_configuration(
            expertise=expertise,
            capabilities=capabilities,
            tools=tools,
            base_config=configuration
        )

        # Phase 3: Create agent instance
        agent = await self.agent_generator.create_agent(
            config=agent_config,
            expertise=expertise
        )

        # Phase 4: Register and initialize agent
        await self.agent_registry.register_agent(agent)
        await self._initialize_agent(agent)

        # Phase 5: Start performance monitoring
        await self.performance_monitor.start_monitoring(agent.id)

        return agent

    async def create_agent_team(
        self,
        project_requirements: ProjectRequirements,
        team_size: int = 3
    ) -> AgentTeam:
        """Create a coordinated team of specialized agents"""

        # Analyze project requirements
        project_analysis = await self.expertise_engine.analyze_project_requirements(
            requirements=project_requirements
        )

        # Determine optimal team composition
        team_composition = await self._determine_team_composition(
            project_analysis=project_analysis,
            team_size=team_size
        )

        # Create individual agents
        agents = []
        for role_config in team_composition.roles:
            agent = await self.create_agent(
                expertise=role_config.expertise,
                capabilities=role_config.capabilities,
                tools=role_config.tools,
                configuration=role_config.config
            )
            agents.append(agent)

        # Create and configure team
        team = AgentTeam(
            agents=agents,
            coordinator=await self._create_team_coordinator(team_composition),
            communication_protocol=team_composition.communication_protocol
        )

        # Register team
        await self.agent_registry.register_team(team)

        return team
```

#### 2. Expertise Detection Engine

```python
class ExpertiseDetectionEngine:
    """AI-powered expertise detection and matching system"""

    def __init__(self, context7_client: Context7Client):
        self.context7_client = context7_client
        self.domain_classifier = DomainClassifier()
        self.expertise_mapper = ExpertiseMapper()
        self.requirement_analyzer = RequirementAnalyzer()

    async def detect_expertise(
        self,
        task_description: str,
        context: Optional[Dict] = None,
        existing_agents: Optional[List[Agent]] = None
    ) -> ExpertiseProfile:
        """Detect optimal expertise profile for a task"""

        # Phase 1: Task analysis
        task_analysis = await self.requirement_analyzer.analyze(
            description=task_description,
            context=context
        )

        # Phase 2: Domain classification
        primary_domain = await self.domain_classifier.classify_primary_domain(
            task_analysis=task_analysis
        )

        # Phase 3: Expertise mapping
        expertise_mapping = await self.expertise_mapper.map_expertise(
            domain=primary_domain,
            task_complexity=task_analysis.complexity_level,
            required_skills=task_analysis.required_skills
        )

        # Phase 4: Conflict resolution with existing agents
        if existing_agents:
            expertise_mapping = await self._resolve_expertise_conflicts(
                proposed_mapping=expertise_mapping,
                existing_agents=existing_agents
            )

        # Phase 5: Create expertise profile
        expertise_profile = ExpertiseProfile(
            primary_domain=primary_domain,
            secondary_domains=expertise_mapping.secondary_domains,
            required_tools=expertise_mapping.required_tools,
            skill_level=expertise_mapping.skill_level,
            complexity_level=task_analysis.complexity_level,
            confidence_score=expertise_mapping.confidence_score,
            contextual_factors=task_analysis.contextual_factors
        )

        return expertise_profile

    async def analyze_project_requirements(
        self,
        requirements: ProjectRequirements
    ) -> ProjectAnalysis:
        """Comprehensive project requirements analysis"""

        # Decompose project into atomic tasks
        atomic_tasks = await self._decompose_project(requirements)

        # Analyze task dependencies
        task_dependencies = await self._analyze_task_dependencies(atomic_tasks)

        # Identify required expertise domains
        expertise_domains = set()
        for task in atomic_tasks:
            task_expertise = await self.detect_expertise(
                task_description=task.description,
                context={"project_context": requirements.context}
            )
            expertise_domains.update(task_expertise.all_domains)

        # Estimate complexity and effort
        complexity_metrics = await self._calculate_project_complexity(
            tasks=atomic_tasks,
            dependencies=task_dependencies,
            expertise_domains=list(expertise_domains)
        )

        return ProjectAnalysis(
            atomic_tasks=atomic_tasks,
            task_dependencies=task_dependencies,
            required_expertise=list(expertise_domains),
            complexity_metrics=complexity_metrics,
            recommended_team_size=self._calculate_optimal_team_size(complexity_metrics)
        )

    async def _resolve_expertise_conflicts(
        self,
        proposed_mapping: ExpertiseMapping,
        existing_agents: List[Agent]
    ) -> ExpertiseMapping:
        """Resolve expertise conflicts with existing agents"""

        conflicts = []
        for agent in existing_agents:
            conflict_score = self._calculate_expertise_overlap(
                proposed=proposed_mapping,
                existing=agent.expertise_profile
            )
            if conflict_score > 0.7:  # High overlap threshold
                conflicts.append({
                    "agent": agent,
                    "conflict_score": conflict_score,
                    "conflicting_domains": self._identify_conflicting_domains(
                        proposed_mapping, agent.expertise_profile
                    )
                })

        if conflicts:
            # Adjust proposed mapping to reduce conflicts
            adjusted_mapping = await self._adjust_expertise_mapping(
                proposed_mapping=proposed_mapping,
                conflicts=conflicts
            )
            return adjusted_mapping

        return proposed_mapping
```

#### 3. Multi-Agent Orchestrator

```python
class MultiAgentOrchestrator:
    """Intelligent multi-agent coordination and orchestration system"""

    def __init__(self, config: OrchestrationConfig):
        self.config = config
        self.task_queue = TaskQueue()
        self.agent_pool = AgentPool()
        self.load_balancer = LoadBalancer()
        self.communication_hub = CommunicationHub()
        self.conflict_resolver = ConflictResolver()

    async def orchestrate_task(
        self,
        task: ComplexTask,
        available_agents: List[Agent]
    ) -> TaskResult:
        """Orchestrate task execution across multiple agents"""

        # Phase 1: Task decomposition
        subtasks = await self._decompose_task(task)

        # Phase 2: Agent selection and assignment
        agent_assignments = await self._assign_agents_to_subtasks(
            subtasks=subtasks,
            available_agents=available_agents
        )

        # Phase 3: Dependency resolution
        execution_plan = await self._resolve_dependencies(
            subtasks=subtasks,
            assignments=agent_assignments
        )

        # Phase 4: Execution orchestration
        results = await self._execute_orchestrated_plan(execution_plan)

        # Phase 5: Result integration
        final_result = await self._integrate_results(results, task)

        return final_result

    async def _assign_agents_to_subtasks(
        self,
        subtasks: List[Subtask],
        available_agents: List[Agent]
    ) -> List[AgentAssignment]:
        """Intelligently assign agents to subtasks based on expertise and load"""

        assignments = []
        agent_loads = {agent.id: agent.current_load for agent in available_agents}

        for subtask in subtasks:
            # Find best-matching agents for subtask
            candidate_agents = await self._find_candidate_agents(
                subtask=subtask,
                available_agents=available_agents,
                agent_loads=agent_loads
            )

            # Select optimal agent considering expertise and load
            selected_agent = await self._select_optimal_agent(
                candidates=candidate_agents,
                subtask=subtask,
                load_balancing=self.config.enable_load_balancing
            )

            # Create assignment
            assignment = AgentAssignment(
                subtask=subtask,
                agent=selected_agent,
                confidence_score=self._calculate_assignment_confidence(
                    subtask, selected_agent
                ),
                estimated_duration=await self._estimate_task_duration(
                    subtask, selected_agent
                )
            )

            assignments.append(assignment)

            # Update agent load
            agent_loads[selected_agent.id] += subtask.estimated_effort

        return assignments

    async def _execute_orchestrated_plan(
        self,
        execution_plan: ExecutionPlan
    ) -> List[SubtaskResult]:
        """Execute orchestrated plan with intelligent coordination"""

        results = []
        active_executions = {}
        completed_subtasks = set()

        # Execute phases in dependency order
        for phase in execution_plan.phases:
            phase_executions = []

            # Start all independent subtasks in the phase
            for subtask in phase.subtasks:
                if subtask.id not in completed_subtasks:
                    assignment = execution_plan.get_assignment(subtask.id)
                    execution = await self._start_subtask_execution(assignment)
                    active_executions[subtask.id] = execution
                    phase_executions.append(execution)

            # Wait for phase completion
            phase_results = await self._wait_for_phase_completion(phase_executions)
            results.extend(phase_results)
            completed_subtasks.update(result.subtask_id for result in phase_results)

            # Handle inter-phase communication
            if phase < len(execution_plan.phases) - 1:
                await self._facilitate_phase_communication(
                    completed_results=phase_results,
                    next_phase=execution_plan.phases[phase + 1]
                )

        return results
```

#### 4. Performance Monitor

```python
class AgentPerformanceMonitor:
    """Real-time agent performance monitoring and optimization"""

    def __init__(self, config: MonitoringConfig):
        self.config = config
        self.metrics_collector = MetricsCollector()
        self.performance_analyzer = PerformanceAnalyzer()
        self.optimization_engine = OptimizationEngine()
        self.alerting_system = AlertingSystem()

    async def start_monitoring(self, agent_id: str) -> None:
        """Start monitoring a specific agent"""

        # Initialize metrics collection
        await self.metrics_collector.initialize_agent_metrics(agent_id)

        # Set up performance tracking
        tracking_config = PerformanceTrackingConfig(
            agent_id=agent_id,
            metrics=self.config.default_metrics,
            collection_interval=self.config.collection_interval,
            alert_thresholds=self.config.default_alert_thresholds
        )

        await self.metrics_collector.start_tracking(tracking_config)

    async def analyze_agent_performance(
        self,
        agent_id: str,
        time_range: TimeRange
    ) -> PerformanceAnalysis:
        """Analyze agent performance over time range"""

        # Collect metrics
        metrics = await self.metrics_collector.get_agent_metrics(
            agent_id=agent_id,
            time_range=time_range
        )

        # Analyze performance patterns
        analysis = await self.performance_analyzer.analyze(
            metrics=metrics,
            agent_id=agent_id
        )

        # Identify optimization opportunities
        optimizations = await self.optimization_engine.identify_optimizations(
            performance_analysis=analysis,
            agent_id=agent_id
        )

        return PerformanceAnalysis(
            metrics=metrics,
            performance_scores=analysis.scores,
            trends=analysis.trends,
            bottlenecks=analysis.bottlenecks,
            optimization_opportunities=optimizations,
            recommendations=analysis.recommendations
        )

    async def optimize_agent_performance(
        self,
        agent_id: str,
        optimizations: List[Optimization]
    ) -> OptimizationResult:
        """Apply performance optimizations to an agent"""

        results = []

        for optimization in optimizations:
            try:
                # Apply optimization
                result = await self._apply_optimization(agent_id, optimization)
                results.append(result)

                # Monitor optimization impact
                if self.config.monitor_optimization_impact:
                    await self._monitor_optimization_impact(
                        agent_id=agent_id,
                        optimization=optimization,
                        pre_optimization_metrics=result.before_metrics
                    )

            except Exception as e:
                logger.error(f"Failed to apply optimization {optimization.id}: {e}")
                results.append(OptimizationResult(
                    optimization_id=optimization.id,
                    success=False,
                    error=str(e)
                ))

        return OptimizationResult(
            agent_id=agent_id,
            optimizations_applied=len([r for r in results if r.success]),
            total_optimizations=len(optimizations),
            performance_improvement=self._calculate_performance_improvement(results),
            results=results
        )
```

### Configuration and Customization

**Agent Factory Configuration**:
```yaml
# agent-factory-config.yaml
agent_factory:
  # Core settings
  expertise_detection:
    enabled: true
    context7_integration: true
    confidence_threshold: 0.8
    max_concurrent_detections: 10

  agent_generation:
    auto_optimization: true
    performance_monitoring: true
    dynamic_scaling: true
    resource_limits:
      max_memory_per_agent: "1GB"
      max_cpu_per_agent: "2.0"
      max_execution_time: 3600

  orchestration:
    load_balancing: true
    conflict_resolution: true
    dependency_management: true
    communication_protocol: "message_queue"

  performance_monitoring:
    enabled: true
    metrics_collection_interval: 60
    optimization_threshold: 0.1
    auto_optimization: true

# Expertise domains configuration
expertise_domains:
  - name: "backend_development"
    primary_skills: ["api_design", "database_modeling", "authentication"]
    secondary_skills: ["performance_optimization", "security"]
    tools: ["fastapi", "django", "postgresql", "redis"]
    complexity_factors: ["concurrent_users", "data_volume", "response_time"]

  - name: "frontend_development"
    primary_skills: ["ui_design", "state_management", "user_experience"]
    secondary_skills: ["performance_optimization", "accessibility"]
    tools: ["react", "vue", "typescript", "webpack"]
    complexity_factors: ["component_count", "state_complexity", "browser_compatibility"]

  - name: "data_science"
    primary_skills: ["statistical_analysis", "machine_learning", "data_visualization"]
    secondary_skills: ["feature_engineering", "model_optimization"]
    tools: ["pandas", "scikit-learn", "tensorflow", "matplotlib"]
    complexity_factors: ["data_size", "model_complexity", "feature_count"]
```

**Advanced Agent Creation**:
```python
# Create specialized agent with custom configuration
custom_agent_config = AgentConfig(
    name="financial_analyst_pro",
    expertise_level="expert",
    capabilities=[
        "data_analysis",
        "financial_modeling",
        "risk_assessment",
        "report_generation",
        "market_prediction"
    ],
    tools=[
        "pandas",
        "numpy",
        "yfinance",
        "matplotlib",
        "seaborn",
        "scikit-learn"
    ],
    constraints={
        "max_execution_time": 1800,  # 30 minutes
        "memory_limit": "2GB",
        "allowed_data_sources": ["yfinance", "alpha_vantage", "fred"]
    },
    performance_targets={
        "accuracy_threshold": 0.95,
        "response_time_max": 10,
        "success_rate_min": 0.90
    },
    learning_enabled=True,
    auto_optimization=True
)

# Create the agent
financial_agent = await agent_factory.create_agent(
    expertise=ExpertiseProfile(
        primary_domain="financial_analysis",
        secondary_domains=["data_science", "risk_management"],
        required_tools=custom_agent_config.tools,
        skill_level="expert"
    ),
    capabilities=custom_agent_config.capabilities,
    tools=custom_agent_config.tools,
    configuration=custom_agent_config
)
```

## Advanced Patterns

### 1. Dynamic Agent Scaling

```python
class DynamicAgentScaler:
    """Intelligent agent scaling based on workload and performance"""

    def __init__(self, agent_factory: AgentFactory):
        self.agent_factory = agent_factory
        self.workload_monitor = WorkloadMonitor()
        self.scaling_engine = ScalingEngine()

    async def monitor_and_scale(self):
        """Continuously monitor workload and scale agents accordingly"""

        while True:
            # Analyze current workload
            workload_analysis = await self.workload_monitor.analyze_workload()

            # Determine scaling needs
            scaling_decision = await self.scaling_engine.decide_scaling(
                workload=workload_analysis,
                current_agents=await self.agent_factory.get_active_agents()
            )

            # Execute scaling actions
            if scaling_decision.scale_up_needed:
                await self._scale_up_agents(scaling_decision.agents_to_add)
            elif scaling_decision.scale_down_needed:
                await self._scale_down_agents(scaling_decision.agents_to_remove)

            # Wait for next monitoring cycle
            await asyncio.sleep(self.monitoring_interval)

    async def _scale_up_agents(self, agent_configs: List[AgentConfig]):
        """Scale up by adding new agents"""

        for config in agent_configs:
            agent = await self.agent_factory.create_agent(
                expertise=config.expertise,
                capabilities=config.capabilities,
                tools=config.tools,
                configuration=config
            )

            await self.agent_factory.deploy_agent(agent)
            logger.info(f"Scaled up: Created {agent.name} agent")

    async def _scale_down_agents(self, agent_ids: List[str]):
        """Scale down by removing idle agents"""

        for agent_id in agent_ids:
            await self.agent_factory.graceful_shutdown_agent(agent_id)
            logger.info(f"Scaled down: Shutdown agent {agent_id}")
```

### 2. Expertise Evolution Engine

```python
class ExpertiseEvolutionEngine:
    """System for evolving agent expertise based on performance and feedback"""

    def __init__(self):
        self.performance_analyzer = AgentPerformanceAnalyzer()
        self.expertise_mutator = ExpertiseMutator()
        self.learning_engine = MachineLearningEngine()

    async def evolve_agent_expertise(
        self,
        agent: Agent,
        performance_history: List[TaskPerformance],
        feedback_data: List[UserFeedback]
    ) -> EvolvedAgent:
        """Evolve agent expertise based on performance and feedback"""

        # Analyze performance patterns
        performance_patterns = await self.performance_analyzer.identify_patterns(
            performance_history
        )

        # Identify expertise gaps and strengths
        expertise_analysis = await self._analyze_expertise_gaps(
            current_expertise=agent.expertise_profile,
            performance_patterns=performance_patterns,
            feedback_data=feedback_data
        )

        # Generate expertise mutations
        mutations = await self.expertise_mutator.generate_mutations(
            current_expertise=agent.expertise_profile,
            gaps=expertise_analysis.gaps,
            strengths=expertise_analysis.strengths
        )

        # Apply successful mutations
        evolved_expertise = await self._apply_expertise_mutations(
            base_expertise=agent.expertise_profile,
            mutations=mutations
        )

        # Create evolved agent
        evolved_agent = await self._create_evolved_agent(
            original_agent=agent,
            evolved_expertise=evolved_expertise
        )

        return evolved_agent

    async def _apply_expertise_mutations(
        self,
        base_expertise: ExpertiseProfile,
        mutations: List[ExpertiseMutation]
    ) -> ExpertiseProfile:
        """Apply successful expertise mutations"""

        evolved_expertise = base_expertise.copy()

        for mutation in mutations:
            if mutation.success_probability > 0.7:  # High confidence mutations
                if mutation.mutation_type == "add_skill":
                    evolved_expertise.required_skills.append(mutation.skill)
                elif mutation.mutation_type == "upgrade_tool":
                    if mutation.tool in evolved_expertise.required_tools:
                        index = evolved_expertise.required_tools.index(mutation.tool)
                        evolved_expertise.required_tools[index] = mutation.upgraded_tool
                elif mutation.mutation_type == "adjust_complexity":
                    evolved_expertise.complexity_level = mutation.new_complexity_level

        return evolved_expertise
```

### 3. Cross-Agent Learning System

```python
class CrossAgentLearningSystem:
    """System for sharing knowledge and learning across multiple agents"""

    def __init__(self):
        self.knowledge_sharer = KnowledgeSharer()
        self.pattern_detector = PatternDetector()
        self.learning_coordinator = LearningCoordinator()

    async def facilitate_cross_agent_learning(
        self,
        agent_community: List[Agent],
        recent_tasks: List[CompletedTask]
    ) -> LearningInsights:
        """Facilitate learning across the agent community"""

        # Identify successful patterns across agents
        successful_patterns = await self.pattern_detector.identify_successful_patterns(
            completed_tasks=recent_tasks,
            agent_community=agent_community
        )

        # Extract shareable knowledge
        shareable_knowledge = await self.knowledge_sharer.extract_shareable_knowledge(
            patterns=successful_patterns,
            agents=agent_community
        )

        # Coordinate knowledge transfer
        learning_results = await self.learning_coordinator.coordinate_knowledge_transfer(
            knowledge=shareable_knowledge,
            recipient_agents=agent_community,
            transfer_strategy="selective"
        )

        return LearningInsights(
            patterns_discovered=len(successful_patterns),
            knowledge_items_shared=len(shareable_knowledge),
            successful_transfers=len([r for r in learning_results if r.successful]),
            learning_efficiency=self._calculate_learning_efficiency(learning_results)
        )

    async def _coordinate_knowledge_transfer(
        self,
        knowledge: List[ShareableKnowledge],
        recipient_agents: List[Agent],
        transfer_strategy: str
    ) -> List[KnowledgeTransferResult]:
        """Coordinate selective knowledge transfer to agents"""

        transfer_results = []

        for agent in recipient_agents:
            # Determine relevant knowledge for this agent
            relevant_knowledge = await self._filter_relevant_knowledge(
                agent=agent,
                all_knowledge=knowledge
            )

            # Transfer knowledge
            for knowledge_item in relevant_knowledge:
                transfer_result = await self._transfer_knowledge_to_agent(
                    agent=agent,
                    knowledge=knowledge_item
                )
                transfer_results.append(transfer_result)

        return transfer_results
```

### 4. Intelligent Task-Agent Matching

```python
class IntelligentTaskAgentMatcher:
    """Advanced task-agent matching using machine learning"""

    def __init__(self):
        self.feature_extractor = TaskAgentFeatureExtractor()
        self.ml_model = TaskAgentMatchingModel()
        self.matching_optimizer = MatchingOptimizer()

    async def find_optimal_agent_match(
        self,
        task: ComplexTask,
        candidate_agents: List[Agent],
        historical_matches: List[HistoricalMatch]
    ) -> AgentMatch:
        """Find optimal agent for task using ML-based matching"""

        # Extract features from task and agents
        task_features = await self.feature_extractor.extract_task_features(task)
        agent_features = await self.feature_extractor.extract_agent_features(
            candidate_agents
        )

        # Generate match scores using ML model
        match_scores = await self.ml_model.predict_matches(
            task_features=task_features,
            agent_features=agent_features,
            historical_matches=historical_matches
        )

        # Optimize matching considering multiple factors
        optimal_match = await self.matching_optimizer.optimize_match(
            match_scores=match_scores,
            constraints={
                "max_response_time": task.deadline,
                "required_expertise_level": task.min_skill_level,
                "resource_availability": self._get_agent_availability(candidate_agents)
            }
        )

        return AgentMatch(
            task=task,
            selected_agent=optimal_match.agent,
            match_score=optimal_match.score,
            confidence_level=optimal_match.confidence,
            estimated_success_probability=optimal_match.success_probability,
            reasoning=optimal_match.reasoning
        )

    async def _optimize_match(
        self,
        match_scores: List[MatchScore],
        constraints: Dict[str, Any]
    ) -> OptimalMatch:
        """Optimize agent selection considering multiple constraints"""

        # Filter agents based on constraints
        viable_agents = []
        for score in match_scores:
            agent = score.agent

            # Check availability
            if not self._is_agent_available(agent, constraints):
                continue

            # Check expertise level
            if agent.expertise_level < constraints["required_expertise_level"]:
                continue

            # Check response time capability
            if agent.avg_response_time > constraints["max_response_time"]:
                continue

            viable_agents.append(score)

        # Select best match from viable agents
        if viable_agents:
            optimal_agent = max(viable_agents, key=lambda x: x.overall_score)

            return OptimalMatch(
                agent=optimal_agent.agent,
                score=optimal_agent.overall_score,
                confidence=optimal_agent.confidence,
                success_probability=optimal_agent.predicted_success_rate,
                reasoning=self._generate_match_reasoning(optimal_agent)
            )
        else:
            # No viable agents found
            return OptimalMatch(
                agent=None,
                score=0.0,
                confidence=0.0,
                success_probability=0.0,
                reasoning="No agents meet the required constraints"
            )
```

### 5. Agent Collaboration Framework

```python
class AgentCollaborationFramework:
    """Framework for enabling intelligent agent collaboration"""

    def __init__(self):
        self.collaboration_matcher = CollaborationMatcher()
        self.communication_protocols = CommunicationProtocolManager()
        self.coordination_engine = CoordinationEngine()
        self.conflict_resolver = ConflictResolver()

    async def enable_agent_collaboration(
        self,
        task: ComplexTask,
        participating_agents: List[Agent]
    ) -> CollaborationResult:
        """Enable and manage agent collaboration for complex tasks"""

        # Phase 1: Task decomposition for collaboration
        collaboration_tasks = await self._decompose_for_collaboration(task)

        # Phase 2: Agent role assignment
        role_assignments = await self.collaboration_matcher.assign_collaboration_roles(
            tasks=collaboration_tasks,
            agents=participating_agents
        )

        # Phase 3: Establish communication protocols
        communication_setup = await self.communication_protocols.setup_collaboration(
            agents=participating_agents,
            roles=role_assignments,
            task_complexity=task.complexity_level
        )

        # Phase 4: Initialize coordination
        coordination_context = await self.coordination_engine.initialize_coordination(
            task=task,
            agents=participating_agents,
            roles=role_assignments,
            communication_setup=communication_setup
        )

        # Phase 5: Execute collaboration
        collaboration_result = await self._execute_collaboration(
            coordination_context=coordination_context
        )

        return collaboration_result

    async def _execute_collaboration(
        self,
        coordination_context: CoordinationContext
    ) -> CollaborationResult:
        """Execute the collaborative task with real-time coordination"""

        active_tasks = coordination_context.tasks
        agent_states = {agent.id: AgentState.IDLE for agent in coordination_context.agents}
        task_results = []

        while active_tasks or any(state == AgentState.WORKING for state in agent_states.values()):
            # Assign tasks to available agents
            for task in active_tasks[:]:
                available_agents = [
                    agent for agent in coordination_context.agents
                    if agent_states[agent.id] == AgentState.IDLE and
                       self._can_agent_handle_task(agent, task)
                ]

                if available_agents:
                    selected_agent = await self._select_best_agent_for_task(
                        task=task,
                        candidates=available_agents
                    )

                    # Assign task
                    await self._assign_task_to_agent(
                        agent=selected_agent,
                        task=task,
                        coordination_context=coordination_context
                    )

                    agent_states[selected_agent.id] = AgentState.WORKING
                    active_tasks.remove(task)

            # Monitor agent progress and handle completions
            completed_tasks = await self._monitor_agent_progress(
                coordination_context=coordination_context,
                agent_states=agent_states
            )

            for completed_task in completed_tasks:
                task_results.append(completed_task.result)
                agent_states[completed_task.agent_id] = AgentState.IDLE

            # Handle conflicts and communication
            await self._handle_collaboration_events(
                coordination_context=coordination_context,
                agent_states=agent_states
            )

            # Wait before next iteration
            await asyncio.sleep(0.1)

        # Finalize collaboration
        final_result = await self._finalize_collaboration(
            task_results=task_results,
            coordination_context=coordination_context
        )

        return CollaborationResult(
            success=final_result.success,
            final_output=final_result.output,
            agent_contributions=self._calculate_agent_contributions(
                task_results, coordination_context.agents
            ),
            collaboration_efficiency=self._calculate_collaboration_efficiency(
                task_results, coordination_context
            ),
            conflicts_resolved=coordination_context.conflicts_resolved,
            lessons_learned=self._extract_collaboration_lessons(
                coordination_context
            )
        )
```

---

## Context7 Integration for Expertise Detection

**Essential Mappings for Agent Creation**:

```python
AGENT_EXPERTISE_MAPPINGS = {
    # Backend Development
    "backend_development": "/tiangolo/fastapi",
    "api_design": "/oauthjs/oauth2-server",
    "database_design": "/mongodb/mongo",
    "authentication": "/openid/connect",
    "microservices": "/microservices/microservices",

    # Frontend Development
    "frontend_development": "/facebook/react",
    "ui_ux_design": "/figma/figma-plugin-docs",
    "state_management": "/reduxjs/redux",
    "web_performance": "/webkit/webkit",

    # Data Science
    "machine_learning": "/tensorflow/tensorflow",
    "data_analysis": "/pandas-dev/pandas",
    "statistical_analysis": "/numpy/numpy",
    "data_visualization": "/matplotlib/matplotlib",

    # DevOps
    "containerization": "/docker/docker",
    "ci_cd": "/actions/toolkit",
    "infrastructure": "/hashicorp/terraform",
    "monitoring": "/prometheus/prometheus",

    # Security
    "cybersecurity": "/owasp/owasp-top-ten",
    "cryptography": "/openssl/openssl",
    "secure_coding": "/sonarsource/sonarqube"
}
```

---

## Integration Patterns

**Microservices Architecture Integration**:
```python
# Agent Factory as a microservice
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

app = FastAPI(title="Agent Factory API")

class AgentCreationRequest(BaseModel):
    task_description: str
    capabilities: List[str]
    tools: List[str]
    configuration: Optional[Dict] = None

class AgentResponse(BaseModel):
    agent_id: str
    expertise_profile: Dict
    capabilities: List[str]
    creation_timestamp: datetime

@app.post("/api/agents/create", response_model=AgentResponse)
async def create_agent(request: AgentCreationRequest):
    """Create a specialized agent via REST API"""

    try:
        # Detect expertise
        expertise = await agent_factory.detect_expertise(
            task_description=request.task_description
        )

        # Create agent
        agent = await agent_factory.create_agent(
            expertise=expertise,
            capabilities=request.capabilities,
            tools=request.tools,
            configuration=request.configuration
        )

        return AgentResponse(
            agent_id=agent.id,
            expertise_profile=expertise.dict(),
            capabilities=agent.capabilities,
            creation_timestamp=datetime.utcnow()
        )

    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/api/agents/orchestrate")
async def orchestrate_agents(
    task: ComplexTask,
    agent_ids: List[str]
):
    """Orchestrate multiple agents for complex task"""

    agents = [await agent_factory.get_agent(agent_id) for agent_id in agent_ids]
    result = await orchestrator.orchestrate_task(task, agents)
    return result
```

---

## Quick Reference Summary

**Core Capabilities**: Intelligent agent generation, expertise detection, multi-agent orchestration, performance optimization, dynamic scaling

**Key Classes**: `AgentFactory`, `ExpertiseDetectionEngine`, `MultiAgentOrchestrator`, `AgentPerformanceMonitor`, `DynamicAgentScaler`

**Essential Methods**: `create_agent()`, `detect_expertise()`, `orchestrate_task()`, `analyze_agent_performance()`, `scale_agents()`

**Integration Ready**: REST APIs, microservices, event-driven architecture, real-time monitoring, auto-scaling

**Enterprise Features**: Cross-agent learning, expertise evolution, intelligent task-agent matching, collaboration framework, comprehensive performance analytics

**Advanced Patterns**: Dynamic scaling based on workload, expertise evolution through machine learning, intelligent collaboration protocols, context7-powered expertise mapping