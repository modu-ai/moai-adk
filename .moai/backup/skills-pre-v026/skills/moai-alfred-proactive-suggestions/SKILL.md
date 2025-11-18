---
name: "moai-alfred-proactive-suggestions"
version: "4.0.0"
created: 2025-11-11
updated: 2025-11-13
status: stable
description: Enterprise Alfred Proactive Suggestions with AI-powered intelligent assistance, Context7 integration, and intelligent recommendation orchestration for enhanced productivity
keywords: ['alfred-suggestions', 'proactive-assistance', 'intelligent-recommendations', 'context-aware', 'context7-integration', 'ai-orchestration', 'productivity-enhancement']
allowed-tools: 
  - Read
  - Bash
  - Write
  - Edit
  - Glob
  - Grep
  - WebFetch
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# Enterprise Alfred Proactive Suggestions Expert v4.0.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-alfred-proactive-suggestions |
| **Version** | 4.0.0 (2025-11-13) |
| **Tier** | Enterprise Alfred Intelligence Expert |
| **AI-Powered** | ✅ Context7 Integration, Intelligent Suggestions |
| **Auto-load** | On demand when Alfred assistance keywords detected |

---

## What It Does

Enterprise Alfred Proactive Suggestions expert with AI-powered intelligent assistance, Context7 integration, and intelligent recommendation orchestration for enhanced developer productivity and workflow optimization.

**Revolutionary v4.0.0 capabilities**:
- 🤖 **AI-Powered Context Analysis** using Context7 MCP for latest productivity patterns
- 📊 **Intelligent Suggestion Engine** with automated workflow optimization recommendations
- 🚀 **Advanced Proactive Assistance** with AI-driven context-aware help and guidance
- 🔗 **Enterprise Integration Framework** with zero-configuration workflow enhancement
- 📈 **Predictive Productivity Analytics** with usage forecasting and optimization insights

---

## When to Use

**Automatic triggers**:
- Alfred workflow optimization and productivity enhancement discussions
- Context-aware assistance and intelligent help system planning
- Developer workflow analysis and optimization strategy
- Proactive recommendation engine implementation

**Manual invocation**:
- Designing intelligent Alfred assistance architectures
- Implementing proactive suggestion systems for productivity
- Planning context-aware help and guidance systems
- Optimizing developer workflows and team productivity

---

# Quick Reference (Level 1)

## Proactive Suggestions Framework (November 2025)

### Core Components
- **Context Analysis**: Real-time analysis of developer activities and patterns
- **Suggestion Engine**: AI-powered recommendation system based on context
- **Workflow Optimization**: Automated workflow improvement suggestions
- **Help System**: Context-aware help and guidance delivery
- **Productivity Analytics**: Usage pattern analysis and optimization

### Suggestion Types
- **Code Assistance**: Intelligent code completion and refactoring suggestions
- **Tool Recommendations**: Optimal tool suggestions for specific tasks
- **Workflow Improvements**: Process optimization and automation suggestions
- **Learning Resources**: Targeted learning material and documentation
- **Best Practices**: Industry-standard patterns and compliance suggestions

### Integration Points
- **Development Environment**: IDE integration and real-time analysis
- **Version Control**: Git workflow optimization and collaboration
- **Build Systems**: Build optimization and dependency management
- **Documentation**: Automatic documentation generation and maintenance
- **Testing**: Test coverage improvement and automation suggestions

### Intelligence Features
- **Pattern Recognition**: Identify recurring patterns and inefficiencies
- **Learning Adaptation**: Adapt suggestions based on user behavior
- **Team Collaboration**: Suggest team-wide optimizations
- **Compliance Monitoring**: Ensure adherence to coding standards
- **Performance Optimization**: Identify performance bottlenecks and solutions

---

# Core Implementation (Level 2)

## Proactive Suggestions Architecture Intelligence

```python
# AI-powered proactive suggestions architecture optimization with Context7
class ProactiveSuggestionsArchitectOptimizer:
    def __init__(self):
        self.context7_client = Context7Client()
        self.productivity_analyzer = ProductivityAnalyzer()
        self.suggestion_engine = SuggestionEngine()
    
    async def design_optimal_suggestions_architecture(self, 
                                                     requirements: ProductivityRequirements) -> ProactiveSuggestionsArchitecture:
        """Design optimal proactive suggestions architecture using AI analysis."""
        
        # Get latest productivity and AI assistance documentation via Context7
        productivity_docs = await self.context7_client.get_library_docs(
            context7_library_id='/productivity/docs',
            topic="developer productivity workflow optimization 2025",
            tokens=3000
        )
        
        ai_docs = await self.context7_client.get_library_docs(
            context7_library_id='/ai-assistance/docs',
            topic="intelligent suggestions context awareness 2025",
            tokens=2000
        )
        
        # Optimize suggestion engine
        suggestion_configuration = self.suggestion_engine.optimize_suggestions(
            requirements.development_patterns,
            requirements.team_collaboration,
            productivity_docs
        )
        
        # Analyze productivity patterns
        productivity_analysis = self.productivity_analyzer.analyze_patterns(
            requirements.current_workflows,
            requirements.productivity_goals,
            ai_docs
        )
        
        return ProactiveSuggestionsArchitecture(
            suggestion_engine=suggestion_configuration,
            context_analysis=self._design_context_analysis(requirements),
            workflow_optimization=productivity_analysis,
            learning_system=self._implement_learning_system(requirements),
            integration_framework=self._design_integration_framework(requirements),
            monitoring_dashboard=self._create_monitoring_dashboard()
        )
```

## Advanced Suggestion Engine Implementation

```typescript
// Enterprise proactive suggestions with TypeScript
interface SuggestionContext {
  userId: string;
  projectType: string;
  currentActivity: string;
  codebaseContext: CodebaseContext;
  teamContext: TeamContext;
  performanceMetrics: PerformanceMetrics;
}

interface Suggestion {
  id: string;
  type: SuggestionType;
  title: string;
  description: string;
  priority: Priority;
  actionability: Actionability;
  context: SuggestionContext;
  implementation?: ImplementationGuide;
  confidence: number;
  timestamp: Date;
}

export class ProactiveSuggestionEngine {
  private contextAnalyzer: ContextAnalyzer;
  private patternRecognizer: PatternRecognizer;
  private learningAdaptation: LearningAdaptation;
  private suggestionHistory: Map<string, Suggestion[]> = new Map();

  constructor() {
    this.contextAnalyzer = new ContextAnalyzer();
    this.patternRecognizer = new PatternRecognizer();
    this.learningAdaptation = new LearningAdaptation();
  }

  async generateSuggestions(context: SuggestionContext): Promise<Suggestion[]> {
    try {
      // Analyze current context
      const contextAnalysis = await this.contextAnalyzer.analyzeContext(context);
      
      // Recognize patterns
      const patterns = await this.patternRecognizer.recognizePatterns(
        context.codebaseContext,
        context.currentActivity
      );
      
      // Generate suggestions based on analysis
      const suggestions = await this.generateSuggestionsFromAnalysis(
        contextAnalysis,
        patterns,
        context
      );
      
      // Apply learning adaptation
      const adaptedSuggestions = await this.learningAdaptation.adaptSuggestions(
        suggestions,
        context.userId,
        context.teamContext
      );
      
      // Filter and prioritize suggestions
      const prioritizedSuggestions = this.prioritizeSuggestions(adaptedSuggestions);
      
      // Store suggestion history
      this.storeSuggestionHistory(context.userId, prioritizedSuggestions);
      
      return prioritizedSuggestions;
    } catch (error) {
      console.error('Error generating suggestions:', error);
      return [];
    }
  }

  private async generateSuggestionsFromAnalysis(
    contextAnalysis: ContextAnalysis,
    patterns: Pattern[],
    context: SuggestionContext
  ): Promise<Suggestion[]> {
    const suggestions: Suggestion[] = [];

    // Code quality suggestions
    const codeSuggestions = await this.generateCodeQualitySuggestions(
      contextAnalysis.codeAnalysis,
      patterns
    );
    suggestions.push(...codeSuggestions);

    // Performance optimization suggestions
    const performanceSuggestions = await this.generatePerformanceSuggestions(
      contextAnalysis.performanceAnalysis,
      context.performanceMetrics
    );
    suggestions.push(...performanceSuggestions);

    // Workflow optimization suggestions
    const workflowSuggestions = await this.generateWorkflowSuggestions(
      contextAnalysis.workflowAnalysis,
      context.teamContext
    );
    suggestions.push(...workflowSuggestions);

    // Learning and development suggestions
    const learningSuggestions = await this.generateLearningSuggestions(
      contextAnalysis.skillGapAnalysis,
      patterns
    );
    suggestions.push(...learningSuggestions);

    return suggestions;
  }

  private async generateCodeQualitySuggestions(
    codeAnalysis: CodeAnalysis,
    patterns: Pattern[]
  ): Promise<Suggestion[]> {
    const suggestions: Suggestion[] = [];

    // Check for code smells
    const codeSmells = await this.detectCodeSmells(codeAnalysis);
    for (const smell of codeSmells) {
      suggestions.push(this.createCodeSmellSuggestion(smell));
    }

    // Suggest refactoring opportunities
    const refactoringOps = await this.identifyRefactoringOpportunities(codeAnalysis);
    for (const opportunity of refactoringOps) {
      suggestions.push(this.createRefactoringSuggestion(opportunity));
    }

    // Suggest test improvements
    const testSuggestions = await this.generateTestSuggestions(codeAnalysis);
    suggestions.push(...testSuggestions);

    return suggestions;
  }

  private createCodeSmellSuggestion(codeSmell: CodeSmell): Suggestion {
    return {
      id: `code-smell-${Date.now()}`,
      type: 'CODE_QUALITY',
      title: `Code Smell Detected: ${codeSmell.type}`,
      description: `Found ${codeSmell.type} in ${codeSmell.location}. ${codeSmell.description}`,
      priority: codeSmell.severity === 'high' ? 'HIGH' : 'MEDIUM',
      actionability: 'IMMEDIATE',
      context: codeSmell.context,
      implementation: {
        steps: [
          `Review the ${codeSmell.type} in ${codeSmell.location}`,
          `Apply the recommended refactoring pattern`,
          `Run tests to ensure no regression`,
          `Consider adding unit tests for the refactored code`
        ],
        codeExample: codeSmell.example,
        references: codeSmell.references
      },
      confidence: 0.8,
      timestamp: new Date()
    };
  }

  private createRefactoringSuggestion(opportunity: RefactoringOpportunity): Suggestion {
    return {
      id: `refactor-${Date.now()}`,
      type: 'REFACTORING',
      title: `Refactoring Opportunity: ${opportunity.type}`,
      description: opportunity.description,
      priority: opportunity.impact === 'high' ? 'HIGH' : 'MEDIUM',
      actionability: 'PLANNED',
      context: opportunity.context,
      implementation: {
        steps: opportunity.steps,
        codeExample: opportunity.example,
        estimatedEffort: opportunity.effort,
        expectedBenefits: opportunity.benefits
      },
      confidence: 0.7,
      timestamp: new Date()
    };
  }

  private prioritizeSuggestions(suggestions: Suggestion[]): Suggestion[] {
    return suggestions.sort((a, b) => {
      // Sort by priority first
      const priorityOrder = { 'CRITICAL': 4, 'HIGH': 3, 'MEDIUM': 2, 'LOW': 1 };
      const priorityDiff = priorityOrder[b.priority] - priorityOrder[a.priority];
      
      if (priorityDiff !== 0) return priorityDiff;
      
      // Then sort by actionability
      const actionabilityOrder = { 'IMMEDIATE': 3, 'PLANNED': 2, 'FUTURE': 1 };
      const actionabilityDiff = actionabilityOrder[b.actionability] - actionabilityOrder[a.actionability];
      
      if (actionabilityDiff !== 0) return actionabilityDiff;
      
      // Finally sort by confidence
      return b.confidence - a.confidence;
    });
  }

  private storeSuggestionHistory(userId: string, suggestions: Suggestion[]): void {
    const history = this.suggestionHistory.get(userId) || [];
    history.push(...suggestions);
    
    // Keep only last 100 suggestions per user
    if (history.length > 100) {
      this.suggestionHistory.set(userId, history.slice(-100));
    } else {
      this.suggestionHistory.set(userId, history);
    }
  }

  async getSuggestionFeedback(suggestionId: string, feedback: SuggestionFeedback): Promise<void> {
    // Update suggestion based on feedback
    await this.learningAdaptation.updateWithFeedback(suggestionId, feedback);
  }

  async getPersonalizedSuggestions(userId: string): Promise<Suggestion[]> {
    // Get user's suggestion history
    const history = this.suggestionHistory.get(userId) || [];
    
    // Analyze feedback patterns
    const feedbackAnalysis = await this.analyzeFeedbackPatterns(userId, history);
    
    // Generate personalized suggestions based on patterns
    return await this.generatePersonalizedSuggestions(userId, feedbackAnalysis);
  }
}

// Context analyzer implementation
class ContextAnalyzer {
  async analyzeContext(context: SuggestionContext): Promise<ContextAnalysis> {
    return {
      codeAnalysis: await this.analyzeCodeContext(context.codebaseContext),
      performanceAnalysis: await this.analyzePerformanceContext(context),
      workflowAnalysis: await this.analyzeWorkflowContext(context),
      skillGapAnalysis: await this.analyzeSkillGaps(context),
      teamDynamics: await this.analyzeTeamDynamics(context.teamContext)
    };
  }

  private async analyzeCodeContext(codebaseContext: CodebaseContext): Promise<CodeAnalysis> {
    // Analyze code structure, dependencies, and patterns
    return {
      complexity: this.calculateComplexity(codebaseContext),
      maintainability: this.assessMaintainability(codebaseContext),
      testCoverage: this.calculateTestCoverage(codebaseContext),
      documentation: this.assessDocumentation(codebaseContext),
      dependencies: this.analyzeDependencies(codebaseContext)
    };
  }

  private async analyzePerformanceContext(context: SuggestionContext): Promise<PerformanceAnalysis> {
    // Analyze performance metrics and bottlenecks
    return {
      buildTimes: context.performanceMetrics.buildTimes,
      testExecutionTimes: context.performanceMetrics.testExecutionTimes,
      codeQualityMetrics: context.performanceMetrics.codeQuality,
      resourceUsage: context.performanceMetrics.resourceUsage
    };
  }
}

// Types
enum SuggestionType {
  CODE_QUALITY = 'CODE_QUALITY',
  PERFORMANCE = 'PERFORMANCE',
  WORKFLOW = 'WORKFLOW',
  LEARNING = 'LEARNING',
  REFACTORING = 'REFACTORING',
  TESTING = 'TESTING',
  DOCUMENTATION = 'DOCUMENTATION'
}

enum Priority {
  CRITICAL = 'CRITICAL',
  HIGH = 'HIGH',
  MEDIUM = 'MEDIUM',
  LOW = 'LOW'
}

enum Actionability {
  IMMEDIATE = 'IMMEDIATE',
  PLANNED = 'PLANNED',
  FUTURE = 'FUTURE'
}

interface ImplementationGuide {
  steps: string[];
  codeExample?: string;
  references?: string[];
  estimatedEffort?: string;
  expectedBenefits?: string[];
}

interface SuggestionFeedback {
  suggestionId: string;
  action: 'ACCEPTED' | 'REJECTED' | 'DEFERRED';
  rating: number;
  comments?: string;
  actualEffort?: string;
  outcome?: string;
}
```

## Learning Adaptation System

```python
# Learning adaptation for personalized suggestions
import numpy as np
from typing import Dict, List, Tuple
from datetime import datetime, timedelta

class LearningAdaptation:
    def __init__(self):
        self.user_profiles: Dict[str, UserProfile] = {}
        self.feedback_history: Dict[str, List[SuggestionFeedback]] = {}
        self.pattern_analyzer = PatternAnalyzer()
    
    async def adapt_suggestions(self, 
                              suggestions: List[Suggestion],
                              user_id: str,
                              team_context: TeamContext) -> List[Suggestion]:
        """Adapt suggestions based on user behavior and team context."""
        
        # Get user profile
        user_profile = await self.get_user_profile(user_id)
        
        # Get team patterns
        team_patterns = await self.pattern_analyzer.analyze_team_patterns(
            team_context
        )
        
        # Adapt suggestions
        adapted_suggestions = []
        for suggestion in suggestions:
            adapted_suggestion = await self.adapt_single_suggestion(
                suggestion, user_profile, team_patterns
            )
            
            if adapted_suggestion.confidence > 0.5:  # Threshold for suggesting
                adapted_suggestions.append(adapted_suggestion)
        
        return adapted_suggestions
    
    async def adapt_single_suggestion(self, 
                                    suggestion: Suggestion,
                                    user_profile: UserProfile,
                                    team_patterns: TeamPatterns) -> Suggestion:
        """Adapt a single suggestion based on user and team patterns."""
        
        # Adjust confidence based on user preferences
        adjusted_confidence = self.adjust_confidence(
            suggestion.confidence,
            user_profile.suggestion_preferences.get(suggestion.type, 0.5)
        )
        
        # Adjust priority based on user's current workload
        adjusted_priority = self.adjust_priority(
            suggestion.priority,
            user_profile.current_workload,
            suggestion.actionability
        )
        
        # Customize description based on user experience level
        customized_description = self.customize_description(
            suggestion.description,
            user_profile.experience_level
        )
        
        return {
            **suggestion,
            confidence: adjusted_confidence,
            priority: adjusted_priority,
            description: customized_description
        }
    
    async def update_with_feedback(self, 
                                 suggestion_id: str, 
                                 feedback: SuggestionFeedback):
        """Update learning model based on user feedback."""
        
        # Store feedback
        if feedback.suggestionId not in self.feedback_history:
            self.feedback_history[feedback.suggestionId] = []
        self.feedback_history[feedback.suggestionId].append(feedback)
        
        # Update user preferences
        await self.update_user_preferences(feedback)
        
        # Update suggestion patterns
        await self.update_suggestion_patterns(feedback)
    
    async def get_user_profile(self, user_id: str) -> UserProfile:
        """Get or create user profile."""
        if user_id not in self.user_profiles:
            self.user_profiles[user_id] = UserProfile(
                user_id=user_id,
                experience_level=self.estimate_experience_level(user_id),
                suggestion_preferences={},
                current_workload=self.assess_current_workload(user_id),
                learning_goals=self.identify_learning_goals(user_id),
                collaboration_style=self.assess_collaboration_style(user_id)
            )
        
        return self.user_profiles[user_id]
    
    def estimate_experience_level(self, user_id: str) -> ExperienceLevel:
        """Estimate user's experience level based on behavior patterns."""
        # This would analyze code complexity, commit patterns, etc.
        return ExperienceLevel.INTERMEDIATE  # Placeholder
    
    def assess_current_workload(self, user_id: str) -> WorkloadLevel:
        """Assess user's current workload."""
        # This would analyze commit frequency, PR activity, etc.
        return WorkloadLevel.MEDIUM  # Placeholder
    
    def adjust_confidence(self, 
                         base_confidence: float, 
                         user_preference: float) -> float:
        """Adjust suggestion confidence based on user preference."""
        # Weight base confidence with user preference
        return (base_confidence * 0.7) + (user_preference * 0.3)
    
    def adjust_priority(self, 
                       base_priority: Priority, 
                       workload: WorkloadLevel,
                       actionability: Actionability) -> Priority:
        """Adjust suggestion priority based on current workload."""
        
        if workload == WorkloadLevel.HIGH and actionability == Actionability.IMMEDIATE:
            # Lower priority for immediate actions during high workload
            if base_priority == Priority.HIGH:
                return Priority.MEDIUM
            elif base_priority == Priority.MEDIUM:
                return Priority.LOW
        
        elif workload == WorkloadLevel.LOW:
            # Increase priority during low workload
            if base_priority == Priority.MEDIUM:
                return Priority.HIGH
            elif base_priority == Priority.LOW:
                return Priority.MEDIUM
        
        return base_priority
    
    def customize_description(self, 
                            description: str, 
                            experience_level: ExperienceLevel) -> str:
        """Customize suggestion description based on user experience level."""
        
        if experience_level == ExperienceLevel.BEGINNER:
            return f"{description} This is a good practice to learn early in your development journey."
        elif experience_level == ExperienceLevel.EXPERT:
            return f"{description} Consider mentoring others on this practice."
        else:
            return description

# Types
class UserProfile:
    def __init__(self, user_id: str, experience_level: ExperienceLevel, 
                 suggestion_preferences: Dict[str, float], current_workload: WorkloadLevel,
                 learning_goals: List[str], collaboration_style: str):
        self.user_id = user_id
        self.experience_level = experience_level
        self.suggestion_preferences = suggestion_preferences
        self.current_workload = current_workload
        self.learning_goals = learning_goals
        self.collaboration_style = collaboration_style

class SuggestionFeedback:
    def __init__(self, suggestion_id: str, action: str, rating: int, 
                 comments: str = None, actual_effort: str = None, outcome: str = None):
        self.suggestion_id = suggestion_id
        self.action = action
        self.rating = rating
        self.comments = comments
        self.actual_effort = actual_effort
        self.outcome = outcome
        self.timestamp = datetime.now()
```

---

# Reference & Integration (Level 4)

## API Reference

### Core Proactive Suggestions Operations
- `generate_suggestions(context)` - Generate context-aware suggestions
- `get_personalized_suggestions(user_id)` - Get personalized recommendations
- `provide_feedback(suggestion_id, feedback)` - Provide feedback on suggestions
- `analyze_productivity_patterns(user_id)` - Analyze user productivity patterns
- `optimize_workflow(workflow_context)` - Suggest workflow optimizations

### Context7 Integration
- `get_latest_productivity_docs()` - Productivity patterns via Context7
- `analyze_developer_patterns()` - Developer patterns via Context7
- `optimize_suggestion_engine()` - Suggestion optimization via Context7

## Best Practices (November 2025)

### DO
- Provide context-aware and actionable suggestions
- Learn from user behavior and adapt recommendations
- Balance proactive assistance with user autonomy
- Provide clear implementation guidance and examples
- Consider user's current workload and priorities
- Offer different suggestion types (code, workflow, learning)
- Measure suggestion effectiveness and user satisfaction
- Respect user privacy and provide opt-out options

### DON'T
- Overwhelm users with too many suggestions
- Make suggestions without proper context analysis
- Ignore user feedback and preferences
- Provide generic or non-actionable recommendations
- Interrupt critical development workflows
- Ignore team collaboration dynamics
- Skip performance impact analysis
- Forget to validate suggestion accuracy

## Works Well With

- `moai-alfred-workflow` (Alfred workflow integration)
- `moai-alfred-agent-guide` (Agent assistance patterns)
- `moai-foundation-trust` (User trust and adoption)
- `moai-domain-backend` (Backend optimization suggestions)
- `moai-domain-frontend` (Frontend optimization suggestions)
- `moai-essentials-perf` (Performance optimization)
- `moai-security-api` (Security best practices)
- `moai-domain-testing` (Testing optimization)

## Changelog

- **v4.0.0** (2025-11-13): Complete Enterprise v4.0 rewrite with 40% content reduction, 4-layer Progressive Disclosure structure, Context7 integration, advanced learning adaptation, and intelligent suggestion patterns
- **v2.0.0** (2025-11-11): Complete metadata structure, suggestion patterns, learning system
- **v1.0.0** (2025-11-11): Initial proactive suggestions foundation

---

**End of Skill** | Updated 2025-11-13

## Intelligent Assistance Framework

### AI-Powered Features
- Real-time context analysis and pattern recognition
- Personalized suggestion adaptation based on user behavior
- Machine learning for continuous improvement
- Natural language processing for intelligent assistance

### Productivity Enhancement
- Automated workflow optimization suggestions
- Performance bottleneck identification and resolution
- Learning path recommendations and skill gap analysis
- Team collaboration and communication improvements

---

**End of Enterprise Alfred Proactive Suggestions Expert v4.0.0**
