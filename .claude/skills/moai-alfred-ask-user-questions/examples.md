# Alfred Ask User Questions - Examples

## Real-World Implementation Examples

### Example 1: Feature Specification Clarification

**Context**: User request is ambiguous - "Add dashboard feature"

```typescript
// Alfred spec-builder agent implementation
const answer = await AskUserQuestion({
  questions: [
    {
      question: "What type of dashboard functionality do you need?",
      header: "Dashboard Type",
      options: [
        {
          label: "Analytics Dashboard",
          description: "Data visualization with charts, graphs, and metrics tracking"
        },
        {
          label: "Admin Dashboard",
          description: "User management, system settings, and administrative controls"
        },
        {
          label: "User Profile Dashboard",
          description: "Personal account information, preferences, and activity history"
        }
      ]
    },
    {
      question: "Which data sources should the dashboard display?",
      header: "Data Sources",
      multiSelect: true,
      options: [
        {
          label: "User Activity",
          description: "Login history, feature usage, and engagement metrics"
        },
        {
          label: "System Performance",
          description: "Server metrics, response times, and error rates"
        },
        {
          label: "Business Metrics",
          description: "Revenue, conversion rates, and KPI tracking"
        }
      ]
    }
  ]
});

// Process user selections
const dashboardType = answer["Dashboard Type"];
const dataSources = answer["Data Sources"];

// Create clear SPEC based on answers
const spec = createClarifiedSpec(dashboardType, dataSources);
```

### Example 2: Implementation Approach Selection

**Context**: Multiple valid implementation paths exist

```typescript
// Alfred tdd-implementer agent implementation
const answer = await AskUserQuestion({
  questions: [
    {
      question: "How should we implement the authentication system?",
      header: "Authentication Approach",
      options: [
        {
          label: "JWT Tokens",
          description: "Stateless authentication with access/refresh tokens for scalability"
        },
        {
          label: "Session-Based",
          description: "Server-side sessions with database storage for enhanced security"
        },
        {
          label: "Third-Party Auth",
          description: "OAuth integration with providers like Google, GitHub, or Auth0"
        }
      ]
    },
    {
      question: "What's your timeline and team expertise level?",
      header: "Project Constraints",
      options: [
        {
          label: "Rapid Development",
          description: "Quick implementation with team familiar with the technology"
        },
        {
          label: "Enterprise Security",
          description: "Maximum security with comprehensive audit trails"
        },
        {
          label: "Scalability Focus",
          description: "Built for high traffic and future growth"
        }
      ]
    }
  ]
});

// Implementation based on user choices
if (answer["Authentication Approach"] === "JWT Tokens" && 
    answer["Project Constraints"] === "Rapid Development") {
  return implementJWTWithDefaultLibraries();
}
```

### Example 3: Risky Operation Confirmation

**Context**: Destructive action requires explicit consent

```typescript
// Alfred git-manager agent implementation
const answer = await AskUserQuestion({
  questions: [
    {
      question: "This will delete 15 branches and merge changes to main. Continue?",
      header: "Destructive Operation",
      options: [
        {
          label: "Proceed with Deletion",
          description: "Delete branches and merge changes (CANNOT BE UNDONE)"
        },
        {
          label: "Dry Run Only",
          description: "Show what would be deleted without making changes"
        },
        {
          label: "Cancel Operation",
          description: "Abort the entire process and keep all branches"
        }
      ]
    }
  ]
});

if (answer["Destructive Operation"] === "Proceed with Deletion") {
  // Require additional confirmation for high-risk operations
  const finalConfirmation = await AskUserQuestion({
    questions: [
      {
        question: "FINAL WARNING: This action cannot be undone. Type 'DELETE' to confirm:",
        header: "Final Confirmation"
      }
    ]
  });
  
  if (finalConfirmation["Final Confirmation"] === "DELETE") {
    return executeBranchDeletion();
  }
}
```

### Example 4: Multi-Option Feature Selection

**Context**: User wants to enable features but unclear which ones

```typescript
// Alfred cc-manager agent implementation
const answer = await AskUserQuestion({
  questions: [
    {
      question: "Which Claude Code features would you like to enable for this project?",
      header: "Feature Selection",
      multiSelect: true,
      options: [
        {
          label: "Auto Documentation",
          description: "Generate technical docs from code comments and structure"
        },
        {
          label: "Code Review Automation",
          description: "Automated code quality checks and review suggestions"
        },
        {
          label: "Performance Monitoring",
          description: "Real-time performance metrics and optimization suggestions"
        },
        {
          label: "Security Scanning",
          description: "Automatic vulnerability detection and security recommendations"
        }
      ]
    }
  ]
});

// Configure Claude Code based on selections
const selectedFeatures = answer["Feature Selection"];
configureClaudeCodeFeatures(selectedFeatures);
```

### Example 5: Sequential Questions with Conditional Flow

**Context**: Complex decision with dependencies between choices

```typescript
// Alfred skill-factory agent implementation
const answer = await AskUserQuestion({
  questions: [
    {
      question: "What type of Skill do you want to create?",
      header: "Skill Type",
      options: [
        {
          label: "Domain Expertise",
          description: "Deep knowledge for specific technology or framework"
        },
        {
          label: "Process Guide",
          description: "Step-by-step workflow for complex tasks"
        },
        {
          label: "Integration Pattern",
          description: "Third-party service integration best practices"
        }
      ]
    }
  ]
});

// Follow-up question based on initial choice
let followUpQuestions = [];
if (answer["Skill Type"] === "Domain Expertise") {
  followUpQuestions.push({
    question: "Which technology domain will this Skill cover?",
    header: "Technology Domain",
    options: [
      {
        label: "Frontend Framework",
        description: "React, Vue, Angular, or other UI frameworks"
      },
      {
        label: "Backend Service",
        description: "Node.js, Python, Java, or other server technologies"
      },
      {
        label: "Database System",
        description: "SQL, NoSQL, or other data storage technologies"
      }
    ]
  });
} else if (answer["Skill Type"] === "Process Guide") {
  followUpQuestions.push({
    question: "What process will this Skill guide?",
    header: "Process Type",
    options: [
      {
        label: "Development Workflow",
        description: "Step-by-step development process and best practices"
      },
      {
        label: "Deployment Pipeline",
        description: "CI/CD setup and deployment automation"
      },
      {
        label: "Testing Strategy",
        description: "Test planning, implementation, and automation"
      }
    ]
  });
}

if (followUpQuestions.length > 0) {
  const detailedAnswer = await AskUserQuestion({
    questions: followUpQuestions
  });
  return createSkillSpecification(answer, detailedAnswer);
}
```

### Example 6: Error Recovery and Custom Input Handling

**Context**: User selects "Other" or provides unexpected input

```typescript
// Alfred debug-helper agent implementation
async function getErrorContext() {
  const answer = await AskUserQuestion({
    questions: [
      {
        question: "What type of error are you experiencing?",
        header: "Error Type",
        options: [
          {
            label: "Compilation Error",
            description: "Code fails to compile or build"
          },
          {
            label: "Runtime Error",
            description: "Application crashes during execution"
          },
          {
            label: "Logic Error",
            description: "Code runs but produces incorrect results"
          },
          {
            label: "Other",
            description: "Different type of issue not listed above"
          }
        ]
      }
    ]
  });

  // Handle custom input for "Other" option
  if (answer["Error Type"] === "Other") {
    const customError = await AskUserQuestion({
      questions: [
        {
          question: "Please describe the error type in detail:",
          header: "Custom Error Description"
        }
      ]
    });
    
    // Validate custom input
    if (!isValidErrorDescription(customError["Custom Error Description"])) {
      return getErrorContext(); // Ask again with guidance
    }
    
    return { errorType: "Custom", description: customError["Custom Error Description"] };
  }

  return { errorType: answer["Error Type"], description: null };
}
```

## Best Practices Demonstrated

### 1. Clear Question Wording
- ✅ Specific questions about implementation details
- ✅ Avoid vague "What should we do?" questions
- ✅ Provide context for why information is needed

### 2. Informative Option Descriptions
- ✅ Explain implications and trade-offs
- ✅ Include technical details for informed decisions
- ✅ Mention when operations are destructive or irreversible

### 3. Appropriate Question Grouping
- ✅ Related questions grouped together
- ✅ Logical flow from general to specific
- ✅ Conditional questions based on previous answers

### 4. Error Handling and Validation
- ✅ Custom input validation for "Other" options
- ✅ Retry logic for invalid responses
- ✅ Graceful fallbacks and recovery patterns

### 5. User Experience Considerations
- ✅ Progress indicators for multi-step processes
- ✅ Confirmation for high-risk operations
- ✅ Clear labeling and categorization

---

**Last Updated**: 2025-11-11
**Related Documentation**: [reference.md](reference.md)
