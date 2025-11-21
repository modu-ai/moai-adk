        
        for lang in primary_languages:
            docs = await self.context7_client.get_library_docs(
                context7_library_id=await self._resolve_language_library(lang),
                topic="performance optimization ecosystem best practices 2025",
                tokens=2000
            )
            language_docs[lang] = docs
        
        # Analyze project requirements
        requirement_analysis = self._analyze_requirements(requirements)
        
        # Optimize language combinations
        language_combinations = self._generate_language_combinations(
            requirement_analysis,
            language_docs
        )
        
        # Evaluate performance characteristics
        performance_evaluation = await self.performance_analyzer.evaluate_languages(
            language_combinations,
            requirement_analysis.performance_requirements,
            language_docs
        )
        
        # Check compatibility and integration
        compatibility_analysis = await self.compatibility_checker.check_compatibility(
            language_combinations,
            requirement_analysis.integration_requirements
        )
        
        return LanguageSelection(
            recommended_stack=self._select_optimal_stack(
                language_combinations,
                performance_evaluation,
                compatibility_analysis
            ),
            alternative_stacks=self._identify_alternatives(
                language_combinations,
                performance_evaluation
            ),
            performance_comparison=performance_evaluation,
            compatibility_matrix=compatibility_analysis,
            migration_strategy=self._plan_migration_strategy(requirements),
            risk_assessment=self._assess_language_risks(
                language_combinations,
                compatibility_analysis
            )
        )
```

## Multi-Language Architecture Patterns

```python
class MultiLanguageArchitect:
    def __init__(self):
        self.integration_patterns = IntegrationPatternLibrary()
        self.performance_optimizer = CrossLanguageOptimizer()
    
    def design_multi_language_architecture(self, 
                                          language_selection: LanguageSelection,
                                          system_requirements: SystemRequirements) -> MultiLanguageArchitecture:
        """Design optimized multi-language system architecture."""
        
        # Define service boundaries based on language strengths
        service_boundaries = self._define_service_boundaries(
            language_selection.recommended_stack,
            system_requirements.domain_boundaries
        )
        
        # Design integration patterns
        integration_patterns = self.integration_patterns.select_patterns(
            service_boundaries,
            system_requirements.communication_requirements
        )
        
        # Optimize cross-language performance
        performance_optimization = self.performance_optimizer.optimize_cross_language_performance(
            language_selection.recommended_stack,
            service_boundaries,
            integration_patterns
        )
        
        return MultiLanguageArchitecture(
            service_boundaries=service_boundaries,
            integration_patterns=integration_patterns,
            performance_optimization=performance_optimization,
            deployment_strategy=self._design_deployment_strategy(
                service_boundaries,
                language_selection.recommended_stack
            ),
            monitoring_setup=self._configure_monitoring(
                service_boundaries,
                integration_patterns
            )
        )
    
    def _define_service_boundaries(self, 
                                 recommended_stack: LanguageStack,
                                 domain_boundaries: List[DomainBoundary]) -> List[ServiceDefinition]:
        """Define service boundaries based on language strengths."""
        
        services = []
        
        for domain in domain_boundaries:
            optimal_language = self._select_optimal_language_for_domain(
                domain, recommended_stack
            )
            
            service = ServiceDefinition(
                name=domain.name,
                domain=domain,
                language=optimal_language,
                responsibilities=domain.responsibilities,
                interfaces=self._define_service_interfaces(domain, optimal_language),
                dependencies=self._identify_dependencies(domain, domain_boundaries),
                performance_requirements=domain.performance_requirements
            )
            
            services.append(service)
        
        return services

class IntegrationPatternLibrary:
    def __init__(self):
        self.patterns = {
            'rest_api': RESTAPIPattern(),
            'graphql': GraphQLPattern(),
            'message_queue': MessageQueuePattern(),
            'event_bus': EventBusPattern(),
            'shared_database': SharedDatabasePattern(),
            'grpc': GRPCPattern(),
            'websocket': WebSocketPattern()
        }
    
    def select_patterns(self, 
                       service_boundaries: List[ServiceDefinition],
                       communication_requirements: CommunicationRequirements) -> List[IntegrationPattern]:
        """Select optimal integration patterns for service communication."""
        
        selected_patterns = []
        
        for service in service_boundaries:
            for dependency in service.dependencies:
                pattern = self._select_pattern_for_dependency(
                    service, dependency, communication_requirements
                )
                
                if pattern and pattern not in selected_patterns:
                    selected_patterns.append(pattern)
        
        return selected_patterns
```

## Performance Optimization Strategies

```typescript
// Cross-language performance optimization
export class LanguagePerformanceOptimizer {
  private languageProfiles = new Map<string, LanguageProfile>();

  constructor() {
    this.initializeLanguageProfiles();
  }

  private initializeLanguageProfiles() {
    // Rust profile - systems programming
    this.languageProfiles.set('rust', {
      strengths: ['performance', 'memory_safety', 'concurrency'],
      weaknesses: ['development_speed', 'ecosystem_size'],
      useCases: ['systems_programming', 'high_performance_services', 'cli_tools'],
      benchmarks: {
        cpuIntensive: 95,
        memoryEfficiency: 98,
        developmentSpeed: 60,
        ecosystemMaturity: 75
      }
    });

    // TypeScript profile - web development
    this.languageProfiles.set('typescript', {
      strengths: ['type_safety', 'ecosystem', 'tooling'],
      weaknesses: ['runtime_performance', 'memory_usage'],
      useCases: ['web_apis', 'frontend_development', 'microservices'],
      benchmarks: {
        cpuIntensive: 70,
        memoryEfficiency: 65,
        developmentSpeed: 90,
        ecosystemMaturity: 95
      }
    });

    // Go profile - backend services
    this.languageProfiles.set('go', {
      strengths: ['concurrency', 'deployment', 'simplicity'],
      weaknesses: ['generic_programming', 'error_handling'],
      useCases: ['microservices', 'cli_tools', 'network_services'],
      benchmarks: {
        cpuIntensive: 85,
        memoryEfficiency: 80,
        developmentSpeed: 85,
        ecosystemMaturity: 80
      }
    });
  }

  optimizeLanguageSelection(requirements: ProjectRequirements): LanguageOptimization {
    const languageScores = new Map<string, number>();

    // Score each language against requirements
    for (const [language, profile] of this.languageProfiles) {
      let score = 0;

      // Performance requirements
      if (requirements.performance === 'high') {
        score += profile.benchmarks.cpuIntensive * 0.3;
        score += profile.benchmarks.memoryEfficiency * 0.2;
      } else if (requirements.performance === 'medium') {
        score += profile.benchmarks.cpuIntensive * 0.2;
        score += profile.benchmarks.memoryEfficiency * 0.1;
      }

      // Development speed requirements
      if (requirements.timeline === 'short') {
        score += profile.benchmarks.developmentSpeed * 0.3;
      } else {
        score += profile.benchmarks.developmentSpeed * 0.1;
      }

      // Ecosystem maturity requirements
      if (requirements.complexity === 'high') {
        score += profile.benchmarks.ecosystemMaturity * 0.2;
      } else {
        score += profile.benchmarks.ecosystemMaturity * 0.1;
      }

      languageScores.set(language, score);
    }

    // Sort languages by score
    const sortedLanguages = Array.from(languageScores.entries())
      .sort((a, b) => b[1] - a[1])
      .slice(0, 5); // Top 5 languages

    return {
      primaryRecommendation: sortedLanguages[0][0],
      alternatives: sortedLanguages.slice(1).map(([lang]) => lang),
      scores: Object.fromEntries(languageScores),
      reasoning: this.generateReasoning(sortedLanguages, requirements)
    };
  }

  private generateReasoning(
    sortedLanguages: [string, number][], 
    requirements: ProjectRequirements
  ): string {
    const [primary, score] = sortedLanguages[0];
    const profile = this.languageProfiles.get(primary)!;

    let reasoning = `${primary} is recommended because it excels in `;
    
    if (requirements.performance === 'high') {
      reasoning += `performance (CPU: ${profile.benchmarks.cpuIntensive}%, Memory: ${profile.benchmarks.memoryEfficiency}%)`;
    }
    
    if (requirements.timeline === 'short') {
      reasoning += ` and has fast development speed (${profile.benchmarks.developmentSpeed}%)`;
    }
    
    reasoning += `. It's particularly suited for ${profile.useCases.join(', ')}.`;

    return reasoning;
  }
}
```


# Advanced Implementation (Level 3)




## Reference & Resources

See [reference.md](reference.md) for detailed API reference and official documentation.
