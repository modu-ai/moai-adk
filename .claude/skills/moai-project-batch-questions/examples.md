# Project Batch Questions - Real-World Examples

> **Main Skill**: [SKILL.md](SKILL.md)  
> **API Reference**: [reference.md](reference.md)

---

## Example 1: Project Initialization Flow

### Scenario

**User**: Runs `/alfred:0-project` on a new project

**Traditional flow** (5 interactions):
1. Q: Which language for docs? → A: Korean
2. Q: Agent prompt language? → A: English  
3. Q: Your nickname? → A: GOOS
4. Q: Team mode? → A: Yes
5. Q: Git workflow? → A: Feature branch

**Batch-optimized flow** (2 interactions):
1. **Batch 1**: Language + Agent prompt + Nickname → All answers at once
2. **Batch 2**: Team mode settings (conditional) → All answers at once

### Implementation

```typescript
// Batch 1: Language settings (always shown)
const languageBatch = await AskUserQuestion({
  questions: [
    {
      question: "프로젝트 초기화 및 문서 작성에 사용할 언어를 선택하세요.",
      header: "언어 선택",
      multiSelect: false,
      options: [
        { label: "🇰🇷 한국어", description: "모든 대화와 문서를 한국어로 작성" },
        { label: "🌍 English", description: "All dialogs and documentation in English" }
      ]
    },
    {
      question: "Alfred 하위 에이전트들의 프롬프트 언어를 선택하세요.",
      header: "에이전트 언어",
      multiSelect: false,
      options: [
        { label: "🌐 English (Global)", description: "토큰 사용량 15-20% 감소" },
        { label: "🗣️ 선택 언어 (현지화)", description: "선택한 언어로 로컬 효율성" }
      ]
    },
    {
      question: "대화에서 부를 닉네임을 입력하세요. (최대 20자)",
      header: "닉네임",
      multiSelect: false,
      options: [
        { label: "직접 입력", description: "'기타' 옵션에서 원하는 이름 입력" }
      ]
    }
  ]
});

// Batch 2: Team mode (conditional - only if team mode detected)
if (detectedMode === 'team') {
  const teamBatch = await AskUserQuestion({
    questions: [
      {
        question: "GitHub 저장소의 'Automatically delete head branches' 설정 상태는?",
        header: "GitHub 설정",
        multiSelect: false,
        options: [
          { label: "✅ 이미 활성화됨", description: "PR 병합 후 자동 삭제" },
          { label: "❌ 비활성화됨 (권장: 활성화)", description: "설정에서 확인 필요" }
        ]
      },
      {
        question: "SPEC 문서 생성 시 사용할 Git 워크플로우를 선택하세요.",
        header: "Git 워크플로우",
        multiSelect: false,
        options: [
          { label: "📋 Feature Branch + PR", description: "팀 협업 최적" },
          { label: "🔄 Direct Commit", description: "빠른 프로토타이핑" }
        ]
      }
    ]
  });
}
```

### Result Processing

```typescript
// Process language batch responses
const config = {
  language: {
    conversation_language: languageBatch['언어 선택'] === '🇰🇷 한국어' ? 'ko' : 'en',
    agent_prompt_language: languageBatch['에이전트 언어'] === '🌐 English (Global)' ? 'english' : 'localized',
    conversation_language_name: languageBatch['언어 선택'] === '🇰🇷 한국어' ? '한국어' : 'English'
  },
  user: {
    nickname: languageBatch['닉네임'],
    selected_at: new Date().toISOString()
  }
};

// Process team batch responses (if exists)
if (teamBatch) {
  config.github = {
    auto_delete_branches: teamBatch['GitHub 설정'] === '✅ 이미 활성화됨',
    spec_git_workflow: teamBatch['Git 워크플로우'] === '📋 Feature Branch + PR' ? 'feature_branch' : 'develop_direct',
    checked_at: new Date().toISOString()
  };
}

await saveConfig('.moai/config.json', config);
```

---

## Example 2: Settings Modification Flow

### Scenario

**User**: Runs `/alfred:0-project setting` to modify existing configuration

**Challenge**: User wants to change multiple settings without going through full initialization

### Solution: Targeted Batch Templates

```typescript
// Detect what settings exist and allow targeted updates
const currentConfig = await loadConfig('.moai/config.json');

// Create targeted batch based on user's intent
const settingsBatch = await AskUserQuestion({
  questions: [
    {
      question: "어떤 설정을 수정하시겠습니까?",
      header: "설정 선택",
      multiSelect: true,
      options: [
        { 
          label: "언어 설정", 
          description: `현재: ${currentConfig.language.conversation_language_name}` 
        },
        { 
          label: "사용자 닉네임", 
          description: `현재: ${currentConfig.user.nickname}` 
        },
        { 
          label: "팀 모드 설정", 
          description: currentConfig.mode === 'team' ? '팀 모드 활성화됨' : '개인 모드'
        },
        { 
          label: "보고서 생성", 
          description: `현재: ${currentConfig.report_generation || 'default'}` 
        }
      ]
    }
  ]
});

// Follow-up batches based on selection
if (settingsBatch['설정 선택'].includes('언어 설정')) {
  await executeLanguageBatch(currentConfig);
}
if (settingsBatch['설정 선택'].includes('팀 모드 설정')) {
  await executeTeamModeBatch(currentConfig);
}
```

---

## Example 3: Domain Selection for New Project

### Scenario

**User**: Starting a new full-stack project and needs to select technology domains

### Batch Implementation

```typescript
const domainBatch = await AskUserQuestion({
  questions: [
    {
      question: "이 프로젝트에 포함할 기술 도메인을 선택하세요.",
      header: "기술 도메인",
      multiSelect: true,
      options: [
        {
          label: "Backend API",
          description: "REST/GraphQL API, 서버 로직, 데이터베이스"
        },
        {
          label: "Frontend Web",
          description: "React/Vue/Angular, UI 컴포넌트, 클라이언트 로직"
        },
        {
          label: "Mobile App",
          description: "iOS/Android 앱, React Native, Flutter"
        },
        {
          label: "DevOps/인프라",
          description: "CI/CD, Docker, Kubernetes, 클라우드 배포"
        },
        {
          label: "데이터/분석",
          description: "데이터 처리, ML 파이프라인, 분석 대시보드"
        }
      ]
    },
    {
      question: "주요 개발 언어를 선택하세요.",
      header: "주요 언어",
      multiSelect: false,
      options: [
        { label: "Python", description: "FastAPI, Django, 데이터 과학" },
        { label: "TypeScript", description: "Node.js, React, 현대 웹 개발" },
        { label: "Go", description: "고성능 서비스, 마이크로서비스" },
        { label: "Java", description: "스프링 부트, 엔터프라이즈 애플리케이션" }
      ]
    }
  ]
});

// Process domain selection
const selectedDomains = domainBatch['기술 도메인'];
const primaryLanguage = domainBatch['주요 언어'];

// Configure project based on selections
const projectConfig = {
  domains: selectedDomains,
  primary_language: primaryLanguage,
  recommended_skills: getRecommendedSkills(selectedDomains, primaryLanguage),
  suggested_agents: getSuggestedAgents(selectedDomains)
};
```

---

## Example 4: Error Handling and Recovery

### Scenario

**User**: Cancels mid-batch or provides invalid input

### Robust Error Handling

```typescript
async function executeBatchWithRetry(batchTemplate: BatchTemplate, maxRetries = 2): Promise<ExecutionResult> {
  let attempt = 0;
  
  while (attempt < maxRetries) {
    try {
      const responses = await AskUserQuestion(batchTemplate.questions);
      
      // Validate responses
      const validation = validateBatchResponses(responses, batchTemplate.name);
      if (!validation.isValid) {
        // Show validation errors and retry
        const shouldRetry = await AskUserQuestion({
          questions: [{
            question: `입력값 오류: ${validation.errors.join(', ')}\n\n다시 시도하시겠습니까?`,
            header: "오류 처리",
            multiSelect: false,
            options: [
              { label: "다시 시도", description: "동일한 질문 다시 표시" },
              { label: "건너뛰기", description: "기본값 사용" }
            ]
          }]
        });
        
        if (shouldRetry['오류 처리'] === '건너뛰기') {
          return { success: false, action: 'use_defaults', skipped: true };
        }
        
        attempt++;
        continue;
      }
      
      // Success - process responses
      return { success: true, data: responses };
      
    } catch (error) {
      if (error.message.includes('User cancelled')) {
        return { 
          success: false, 
          error: 'User cancelled',
          action: 'abort_or_partial_save'
        };
      }
      
      attempt++;
      if (attempt >= maxRetries) {
        return { 
          success: false, 
          error: `Failed after ${maxRetries} attempts: ${error.message}`,
          action: 'manual_intervention_required'
        };
      }
    }
  }
  
  return { success: false, error: 'Max retries exceeded', action: 'abort' };
}
```

---

## Example 5: Performance Monitoring

### Scenario

**Alfred**: Wants to track batch performance and UX improvements

### Metrics Collection

```typescript
interface BatchMetrics {
  templateName: string;
  questionCount: number;
  interactionCount: number;
  timeSpent: number;
  userSatisfaction?: number; // Collected via follow-up
  tokenUsage: number;
}

class BatchPerformanceTracker {
  private metrics: BatchMetrics[] = [];
  
  async trackBatchExecution(
    templateName: string, 
    executionFn: () => Promise<any>
  ): Promise<any> {
    const startTime = Date.now();
    const questionCount = this.getQuestionCount(templateName);
    
    try {
      const result = await executionFn();
      const timeSpent = Date.now() - startTime;
      
      // Record successful execution
      this.metrics.push({
        templateName,
        questionCount,
        interactionCount: 1, // Batches always use 1 interaction
        timeSpent,
        tokenUsage: this.estimateTokenUsage(result),
      });
      
      // Ask for satisfaction rating (optional)
      this.requestSatisfactionRating(templateName);
      
      return result;
      
    } catch (error) {
      // Record failed execution
      this.metrics.push({
        templateName,
        questionCount,
        interactionCount: 1,
        timeSpent: Date.now() - startTime,
        tokenUsage: 0,
      });
      
      throw error;
    }
  }
  
  generateReport(): PerformanceReport {
    const totalBatches = this.metrics.length;
    const avgTimeSpent = this.metrics.reduce((sum, m) => sum + m.timeSpent, 0) / totalBatches;
    const totalInteractionsSaved = this.calculateInteractionsSaved();
    
    return {
      totalBatches,
      avgTimeSpent,
      interactionsSaved: totalInteractionsSaved,
      satisfactionScore: this.calculateSatisfactionScore(),
      tokenEfficiency: this.calculateTokenEfficiency()
    };
  }
  
  private calculateInteractionsSaved(): number {
    return this.metrics.reduce((total, metric) => {
      return total + (metric.questionCount - metric.interactionCount);
    }, 0);
  }
}
```

---

## Integration Checklist

### For Each Batch Template

- [ ] **Question grouping**: Related questions batched together
- [ ] **Clear headers**: Short, descriptive headers (≤12 chars)
- [ ] **Concise labels**: 1-5 words per option
- [ ] **Helpful descriptions**: Context for informed choices
- [ ] **Response validation**: Input checking and error handling
- [ ] **Configuration mapping**: Responses → config format
- [ ] **Multi-language support**: Templates for different languages
- [ ] **Conditional logic**: Show/hide based on context
- [ ] **Error recovery**: Graceful handling of failures
- [ ] **Performance tracking**: Metrics collection

### For Integration Points

- [ ] **Command integration**: Alfred commands use batches
- [ ] **Sub-agent support**: Agents can invoke batch templates
- [ ] **Configuration persistence**: Save responses to config files
- [ ] **Backward compatibility**: Support existing response formats
- [ ] **Testing**: Validate with sample responses
- [ ] **Documentation**: Examples and integration guides

---

**End of Examples** | Created 2025-11-05 | Real-world batch implementations
