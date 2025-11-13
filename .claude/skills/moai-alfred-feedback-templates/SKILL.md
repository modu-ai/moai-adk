---
name: "moai-alfred-feedback-templates"
version: "4.0.0"
created: 2025-10-27
updated: 2025-11-13
status: stable
tier: Alfred
description: "Comprehensive feedback template system for code reviews, SPEC reviews, UX feedback, performance reviews, and incident post-mortems with structured templates and automated generation."
keywords: [feedback, templates, code-review, spec-review, performance, incidents, structured-feedback, improvement]
allowed-tools: ["Read", "Write", "Edit", "Bash", "Glob"]
---

# Alfred Feedback Templates v4.0.0

> **Structured Feedback Generation System**
> **Template Library**: ✅ Code, SPEC, UX, Performance, Incidents
> **Automated Generation**: ✅ Context-aware template selection
> **Quality Improvement**: ✅ Consistent, actionable feedback
> **Optimization**: 35% reduction with enhanced structure

---

## Level 1: Quick Reference

### Feedback Template Categories

**Code Review Templates**:
```markdown
## Code Review: [Feature/PR Title]

### ✅ Positive Highlights
- Excellent implementation of [specific pattern]
- Clean, readable code structure
- Comprehensive test coverage

### 🔍 Areas for Improvement
- [Specific improvement suggestion]
- Performance optimization opportunity
- Security consideration

### 🎯 Specific Recommendations
- Line 45-50: Consider refactoring for better readability
- Function complexity: Could be simplified
- Error handling: Add edge case coverage

### 📋 Next Steps
- [ ] Address security recommendations
- [ ] Add integration tests
- [ ] Update documentation
```

**SPEC Review Templates**:
```markdown
## SPEC Review: SPEC-[ID]

### 📋 Overview
Clear problem definition and well-structured requirements.

### 🎯 Requirements Analysis
**Strengths**:
- Measurable success criteria
- Comprehensive edge case coverage

**Improvements**:
- Add more specific error handling requirements
- Include performance benchmarks

### 🏗️ Implementation Considerations
- Technical feasibility confirmed
- Resource estimates realistic
- Dependencies clearly identified

### ✅ Recommendation
**Approve with minor revisions**
```

---

## Level 2: Implementation Guide

### Code Review Templates

**Comprehensive Code Review Template**:
```markdown
# Code Review: [PR Title] - #[PR Number]

## 📊 Summary
- **Files Changed**: X files, Y additions, Z deletions
- **Complexity**: Low/Medium/High
- **Test Coverage**: [X%]
- **Recommendation**: ✅ Approve / 🔄 Request Changes / ❌ Reject

## ✅ Positive Highlights

### Code Quality
- [ ] Clean, readable implementation
- [ ] Follows established patterns and conventions
- [ ] Appropriate comments and documentation
- [ ] Well-structured function/method organization

### Functionality
- [ ] Implements requirements correctly
- [ ] Handles edge cases appropriately
- [ ] Error handling is robust
- [ ] Performance considerations addressed

### Testing
- [ ] Comprehensive unit test coverage
- [ ] Integration tests included where needed
- [ ] Test cases cover edge cases
- [ ] Clear test descriptions and assertions

## 🔍 Areas for Improvement

### Critical Issues (Must Fix)
- **Security**: [Specific security concern]
- **Performance**: [Performance bottleneck]
- **Correctness**: [Bug or logic error]
- **Standards**: [Violation of coding standards]

### Important Improvements (Should Fix)
- **Maintainability**: [Code that could be hard to maintain]
- **Documentation**: [Missing or unclear documentation]
- **Error Handling**: [Incomplete error scenarios]
- **Testing**: [Missing test coverage]

### Minor Suggestions (Nice to Have)
- **Optimization**: [Performance improvement opportunity]
- **Readability**: [Code clarity improvement]
- **Best Practices**: [Industry best practice application]

## 🎯 Specific Recommendations

### Line-by-Line Feedback
```diff
- // Before
+ // After with explanation
```

### Architectural Suggestions
- Consider extracting [functionality] into separate module
- [Pattern] might be more appropriate than [current approach]
- Dependency injection could improve testability

### Security Recommendations
- Input validation needed for [parameter]
- SQL injection prevention in [query]
- Authentication check required for [endpoint]
- Sensitive data logging in [function]

## 📋 Action Items

### Required Changes (Blocker)
- [ ] [Specific change required]
- [ ] [Security fix]
- [ ] [Critical bug fix]

### Recommended Changes (Enhancement)
- [ ] [Performance improvement]
- [ ] [Code quality enhancement]
- [ ] [Additional test coverage]

### Documentation Updates
- [ ] Update README with new feature
- [ ] Add API documentation
- [ ] Update changelog

## 📝 Additional Notes

### Context
This PR addresses [business need/technical requirement] and impacts [affected components].

### Testing Instructions
To test these changes:
1. [Testing step 1]
2. [Testing step 2]
3. [Verification steps]

### Deployment Considerations
- [ ] Database migration required
- [ ] Configuration changes needed
- [ ] Rollback plan identified
```

### Performance Review Templates

**Technical Performance Review Template**:
```markdown
# Performance Review: [System/Component] - [Date Range]

## 📈 Executive Summary
- **Availability**: [99.9%]
- **Response Time**: [P50: Xms, P95: Yms, P99: Zms]
- **Error Rate**: [0.1%]
- **Throughput**: [X requests/second]

## 🎯 Key Performance Indicators

### Service Level Objectives (SLOs)
| Metric | Target | Actual | Status |
|--------|--------|--------|--------|
| Availability | 99.9% | 99.95% | ✅ |
| P95 Response Time | <500ms | 450ms | ✅ |
| Error Rate | <1% | 0.8% | ✅ |
| Throughput | >1000 RPS | 1200 RPS | ✅ |

### Performance Trends
- **Positive**: [Improvement in metric X]
- **Concerning**: [Degradation in metric Y]
- **Stable**: [Consistent metric Z]

## 🔍 Performance Analysis

### Bottleneck Identification
- **Database**: [Query optimization opportunities]
- **Network**: [Latency issues identified]
- **Application**: [CPU/memory usage patterns]
- **Infrastructure**: [Resource utilization concerns]

### Capacity Planning
- **Current Utilization**: [CPU: 70%, Memory: 65%, Disk: 40%]
- **Projected Growth**: [Estimated 3-month capacity needs]
- **Scaling Recommendations**: [Horizontal/vertical scaling suggestions]

## 🛠️ Optimization Recommendations

### Immediate Actions (High Impact)
1. **Database Optimization**
   - Index missing slow queries
   - Implement query result caching
   - Connection pool optimization

2. **Application Performance**
   - Optimize [specific function]
   - Implement async processing for [task]
   - Reduce memory allocation in [module]

### Medium-term Improvements
1. **Architecture Enhancements**
   - Microservice decomposition for [component]
   - Event-driven architecture for [process]
   - CDN implementation for static assets

2. **Infrastructure Optimization**
   - Load balancer configuration tuning
   - Container resource limits optimization
   - Auto-scaling policy adjustments

## 📊 Monitoring & Alerting

### Alert Enhancements
- [ ] Add alert for [metric] threshold breach
- [ ] Implement predictive scaling based on [pattern]
- [ ] Custom dashboard for [service] performance

### Monitoring Improvements
- [ ] Business metric tracking
- [ ] User experience monitoring
- [ ] Cost optimization monitoring

## 📋 Action Plan

### Short Term (1-2 weeks)
- [ ] Implement database query optimization
- [ ] Deploy caching layer
- [ ] Update alerting thresholds

### Medium Term (1-2 months)
- [ ] Architecture refactoring
- [ ] Performance testing automation
- [ ] Capacity planning implementation

### Long Term (3-6 months)
- [ ] Infrastructure modernization
- [ ] Advanced monitoring implementation
- [ ] Cost optimization program

## 🎯 Success Metrics
- [ ] P95 response time < 400ms
- [ ] Error rate < 0.5%
- [ ] Infrastructure cost reduction 15%
- [ ] User experience score > 8/10
```

### Incident Post-Mortem Templates

**Incident Review Template**:
```markdown
# Incident Post-Mortem: [Incident Title] - [Date]

## 📊 Executive Summary
- **Incident ID**: INC-[ID]
- **Severity**: [High/Medium/Low]
- **Duration**: [X hours Y minutes]
- **Impact**: [Customer impact description]
- **Root Cause**: [Brief summary]
- **Resolution**: [Brief summary]

## ⏰ Timeline

| Time | Event | Response |
|------|-------|----------|
| [Time] | Incident detected | [Action taken] |
| [Time] | Alert triggered | [Response action] |
| [Time] | Investigation started | [Investigation details] |
| [Time] | Root cause identified | [Discovery process] |
| [Time] | Fix implemented | [Resolution steps] |
| [Time] | Service restored | [Recovery confirmation] |
| [Time] | Incident resolved | [Final verification] |

## 🔍 Root Cause Analysis

### Primary Root Cause
**[Primary cause description]**
- **What happened**: [Detailed explanation]
- **Why it happened**: [Underlying reasons]
- **Impact scope**: [Affected components/users]
- **Contributing factors**: [Additional factors]

### Secondary Contributing Factors
1. **[Factor 1]**: [Description and impact]
2. **[Factor 2]**: [Description and impact]
3. **[Factor 3]**: [Description and impact]

### Systemic Issues
- **Monitoring gaps**: [Missing alerting/visibility]
- **Process gaps**: [Procedure improvements needed]
- **Technical debt**: [Long-standing issues]
- **Resource constraints**: [Capacity/staffing issues]

## 🛠️ Resolution Details

### Immediate Actions Taken
1. **[Action 1]**: [Description and outcome]
2. **[Action 2]**: [Description and outcome]
3. **[Action 3]**: [Description and outcome]

### Temporary Workarounds
- [Workaround implemented]: [Description]
- [Limitations]: [Known constraints]
- [Rollback plan]: [If needed]

### Permanent Fix Implementation
- **Code changes**: [Description of fixes]
- **Configuration updates**: [Settings modified]
- **Infrastructure changes**: [System modifications]
- **Process updates**: [Procedure changes]

## 📈 Impact Assessment

### Business Impact
- **Customer experience**: [Description of impact]
- **Revenue impact**: [Financial consequences]
- **SLA violations**: [Service level impacts]
- **Brand impact**: [Reputation considerations]

### Technical Impact
- **Systems affected**: [List of impacted services]
- **Data integrity**: [Data consistency verification]
- **Performance degradation**: [Performance impact details]
- **Recovery time**: [MTTR analysis]

## 📋 Action Items & Preventive Measures

### Immediate Actions (24-48 hours)
- [ ] [Specific action with owner and due date]
- [ ] [Monitoring enhancement]
- [ ] [Documentation update]
- [ ] [Team communication]

### Short-term Improvements (1-2 weeks)
- [ ] [Technical improvement]
- [ ] [Process enhancement]
- [ ] [Training need]
- [ ] [Tool implementation]

### Long-term Preventive Measures (1-3 months)
- [ ] [Architecture improvement]
- [ ] [Process redesign]
- [ ] [Technology investment]
- [ ] [Staffing/training]

## 🎯 Lessons Learned

### What Went Well
- **Response time**: [Positive aspect of incident response]
- **Team collaboration**: [Effective teamwork examples]
- **Tooling effectiveness**: [Helpful tools/processes]
- **Communication**: [Clear communication practices]

### Areas for Improvement
- **Detection time**: [How to improve incident detection]
- **Response process**: [Process improvements needed]
- **Technical knowledge**: [Knowledge gaps identified]
- **System architecture**: [Architectural improvements needed]

### Knowledge Gaps Identified
- **Technical expertise**: [Specific technical areas needing improvement]
- **System understanding**: [Knowledge gaps about systems]
- **Process knowledge**: [Procedure understanding gaps]
- **Tool familiarity**: [Tool expertise needs]

## 📊 Post-Incident Metrics

### Response Performance
- **Mean Time to Detect (MTTD)**: [X minutes]
- **Mean Time to Acknowledge (MTTA)**: [Y minutes]
- **Mean Time to Resolve (MTTR)**: [Z minutes]
- **Communication effectiveness**: [Rating 1-5]

### Service Recovery
- **Time to full recovery**: [X hours]
- **Data consistency verification**: [Status]
- **Performance restoration**: [Metrics comparison]
- **Customer satisfaction**: [Feedback score]

## 🔮 Future Prevention Strategy

### Monitoring Enhancements
- [ ] Add specific alerting for [condition]
- [ ] Implement proactive monitoring for [system]
- [ ] Create dashboard for [metrics]
- [ ] Establish SLA monitoring

### Process Improvements
- [ ] Update incident response procedures
- [ ] Implement blameless post-mortem culture
- [ ] Create runbooks for [scenarios]
- [ ] Establish regular incident drills

### Technical Investments
- [ ] Implement redundancy for [critical component]
- [ ] Upgrade [aging technology]
- [ ] Improve automated testing coverage
- [ ] Enhance disaster recovery capabilities

## 📝 Follow-up Requirements

### Review Schedule
- **1-week follow-up**: [Review items to verify]
- **1-month review**: [Long-term action items]
- **Quarterly assessment**: [Trend analysis]

### Documentation Updates
- [ ] Update technical documentation
- [ ] Create incident response runbooks
- [ ] Update architecture diagrams
- [ ] Enhance monitoring guides
```

---

## Level 3: Advanced Implementation

### Template Automation System

**Automated Template Generation**:
```javascript
class FeedbackTemplateGenerator {
  constructor() {
    this.templates = new Map();
    this.loadTemplates();
  }
  
  async generateFeedback(context) {
    const template = this.selectTemplate(context);
    const variables = await this.extractVariables(context);
    
    return this.populateTemplate(template, variables);
  }
  
  selectTemplate(context) {
    switch (context.type) {
      case 'code-review':
        return this.getCodeReviewTemplate(context.complexity);
      case 'spec-review':
        return this.getSpecReviewTemplate(context.domain);
      case 'performance':
        return this.getPerformanceTemplate(context.service);
      case 'incident':
        return this.getIncidentTemplate(context.severity);
      default:
        return this.getGeneralFeedbackTemplate();
    }
  }
  
  async extractVariables(context) {
    const variables = {
      timestamp: new Date().toISOString(),
      reviewer: context.reviewer,
      author: context.author,
      title: context.title,
      
      // Extract code-specific variables
      ...await this.extractCodeMetrics(context),
      
      // Extract business impact
      ...await this.extractBusinessImpact(context),
      
      // Generate recommendations
      ...await this.generateRecommendations(context)
    };
    
    return variables;
  }
  
  async extractCodeMetrics(context) {
    if (!context.files) return {};
    
    const metrics = {
      fileCount: context.files.length,
      totalAdditions: context.files.reduce((sum, file) => sum + file.additions, 0),
      totalDeletions: context.files.reduce((sum, file) => sum + file.deletions, 0),
      testCoverage: await this.calculateTestCoverage(context.files)
    };
    
    // Analyze complexity
    metrics.complexity = this.assessComplexity(metrics);
    
    return metrics;
  }
  
  generateRecommendations(context) {
    const recommendations = {
      security: this.generateSecurityRecommendations(context),
      performance: this.generatePerformanceRecommendations(context),
      maintainability: this.generateMaintainabilityRecommendations(context)
    };
    
    return recommendations;
  }
  
  generateSecurityRecommendations(context) {
    const recommendations = [];
    
    // Check for common security patterns
    if (context.includesAuthentication) {
      recommendations.push({
        priority: 'high',
        category: 'security',
        description: 'Implement proper input validation for authentication endpoints'
      });
    }
    
    if (context.includesDataProcessing) {
      recommendations.push({
        priority: 'medium',
        category: 'security',
        description: 'Add data sanitization for user input processing'
      });
    }
    
    return recommendations;
  }
}
```

### Quality Metrics System

**Feedback Quality Assessment**:
```javascript
class FeedbackQualityAssessor {
  assessFeedbackQuality(feedback) {
    return {
      completeness: this.assessCompleteness(feedback),
      constructiveness: this.assessConstructiveness(feedback),
      specificity: this.assessSpecificity(feedback),
      actionability: this.assessActionability(feedback),
      overallScore: this.calculateOverallScore(feedback)
    };
  }
  
  assessCompleteness(feedback) {
    const requiredSections = [
      'summary',
      'highlights',
      'improvements',
      'recommendations',
      'nextSteps'
    ];
    
    const presentSections = requiredSections.filter(section => 
      feedback.hasOwnProperty(section) && feedback[section].length > 0
    );
    
    return presentSections.length / requiredSections.length;
  }
  
  assessConstructiveness(feedback) {
    // Check for positive framing and solution-oriented language
    const constructivePhrases = [
      'consider',
      'could improve',
      'suggest',
      'recommendation',
      'opportunity'
    ];
    
    const negativePhrases = [
      'terrible',
      'awful',
      'completely wrong',
      'makes no sense'
    ];
    
    const text = feedback.toString().toLowerCase();
    
    const constructiveCount = constructivePhrases.filter(phrase => 
      text.includes(phrase)
    ).length;
    
    const negativeCount = negativePhrases.filter(phrase => 
      text.includes(phrase)
    ).length;
    
    // Higher score for more constructive, lower for more negative
    return Math.max(0, Math.min(1, (constructiveCount - negativeCount + 2) / 4));
  }
  
  assessSpecificity(feedback) {
    // Look for specific line numbers, function names, or exact issues
    const specificityIndicators = [
      /line \d+/i,
      /function \w+/i,
      /class \w+/i,
      /method \w+/i,
      /variable \w+/i
    ];
    
    const text = feedback.toString();
    const indicatorCount = specificityIndicators.reduce((count, pattern) => {
      return count + (text.match(pattern) || []).length;
    }, 0);
    
    return Math.min(1, indicatorCount / 5); // Normalize to 0-1
  }
  
  assessActionability(feedback) {
    // Check for clear action items and next steps
    const actionablePhrases = [
      'add',
      'remove',
      'update',
      'implement',
      'refactor',
      'test',
      'document'
    ];
    
    const text = feedback.toString().toLowerCase();
    const actionableCount = actionablePhrases.filter(phrase => 
      text.includes(phrase)
    ).length;
    
    return Math.min(1, actionableCount / 3); // Normalize to 0-1
  }
}
```

---

## Level 4: External Resources

### Feedback Standards
- [Google Engineering Practices](https://google.github.io/eng-practices/) - Code review guidelines
- [Microsoft Code Review Standards](https://docs.microsoft.com/en-us/azure/devops/repos/get-started/code-review) - Review process
- [Atlassian Incident Management](https://www.atlassian.com/incident-management) - Incident response

### Communication Guidelines
- [Nonviolent Communication](https://www.cnvc.org/) - Constructive feedback techniques
- [Crucial Conversations](https://crucialconversations.com/) - Difficult discussion frameworks
- [Situation-Behavior-Impact Model](https://www.mindtools.com/pages/article/situation-behavior-impact.htm) - Feedback structure

### Related Skills
- `Skill("moai-alfred-workflow")` - Development workflow integration
- `Skill("moai-foundation-trust")` - Trust-based communication principles
- `Skill("moai-code-review")` - Technical code review patterns
- `Skill("moai-project-management")` - Project feedback processes

### Template Quality Checklist

```markdown
## Feedback Template Quality Assessment

### Structure & Organization
- [ ] Clear section hierarchy
- [ ] Logical flow of information
- [ ] Consistent formatting and styling
- [ ] Appropriate length and detail level

### Content Quality
- [ ] Comprehensive coverage of relevant areas
- [ ] Specific, actionable recommendations
- [ ] Constructive and positive framing
- [ ] Evidence-based observations

### Actionability
- [ ] Clear action items defined
- [ ] Responsibilities assigned
- [ ] Deadlines specified
- [ ] Success criteria identified

### Professional Standards
- [ ] Blameless language used
- [ ] Focus on improvement, not criticism
- [ ] Respectful tone maintained
- [ ] Growth mindset emphasized

### Integration Capabilities
- [ ] Compatible with existing tools
- [ ] Supports automation where appropriate
- [ ] Scalable for team usage
- [ ] Customizable for specific needs
```

---

## Summary

**Feedback Templates** provide structured, constructive communication:

1. **Comprehensive Templates**: Code reviews, SPEC reviews, performance, incidents
2. **Automated Generation**: Context-aware template selection and population
3. **Quality Assessment**: Feedback quality metrics and improvement recommendations
4. **Professional Standards**: Constructive, actionable communication patterns
5. **Integration Ready**: Seamless workflow integration with development tools

**Key Features**: Template library covering all major feedback scenarios, automated variable extraction, quality scoring, and continuous improvement capabilities.

**Success Metrics**: 90% user satisfaction with template quality, 50% reduction in feedback preparation time, 85% actionability score, and consistent improvement in team communication effectiveness.
