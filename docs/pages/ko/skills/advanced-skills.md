# Advanced Skills 완전 가이드

## 개요

MoAI-ADK는 8개의 고급 Skills를 통해 전문적인 AI 기반 개발 워크플로우를 제공합니다. Context7 통합, 문서 처리, React 컴포넌트 생성, 엔터프라이즈 커뮤니케이션, 그리고 자동화된 테스팅까지 - 현대적인 애플리케이션 개발에 필요한 모든 전문 기능을 포함합니다.

**핵심 Advanced Skills**:
- **MCP Builder**: Context7 기반 MCP 서버 개발
- **Document Processing**: AI 기반 문서 처리 파이프라인
- **Artifacts Builder**: React + Tailwind 컴포넌트 생성
- **Internal Communications**: 엔터프라이즈 커뮤니케이션
- **Playwright Testing**: 웹 앱 자동화 테스팅
- **Development Chat**: AI 페어 프로그래밍
- **Prompt Engineering**: LLM 프롬프트 최적화
- **Video Analysis**: 영상 콘텐츠 분석

## MCP Builder Skill

### 개요

Model Context Protocol (MCP) 서버를 Context7 표준에 따라 개발하는 전문 Skill입니다. 최신 AI 아키텍처 패턴과 산업 표준을 준수하며, 프로덕션급 MCP 서버를 빠르게 구축합니다.

### 핵심 기능

**Context7 완전 통합**:
```typescript
// MCP 서버 스켈레톤 자동 생성
import { Context7MCP } from '@moai/mcp-builder'

const mcpServer = new Context7MCP({
  name: 'my-data-source',
  version: '1.0.0',
  capabilities: ['query', 'index', 'sync'],
  context7: {
    enabled: true,
    validation: 'strict',
    caching: true,
  },
})

// Context7 표준 엔드포인트 자동 생성
mcpServer.registerResource({
  uri: 'data://my-source/*',
  handler: async (uri, options) => {
    // Context7 validation
    const validated = await context7.validate(uri, options)

    // 데이터 가져오기
    const data = await fetchData(validated.resourceId)

    // Context7 format 변환
    return context7.format(data, {
      maxTokens: options.tokens || 5000,
      format: 'structured',
    })
  },
})
```

**AI-Powered Architecture**:
```typescript
// Skill 호출로 MCP 서버 생성
Skill("moai-advanced-mcp-builder", {
  prompt: `
    Create an MCP server for PostgreSQL database integration:

    Requirements:
    - Connect to PostgreSQL
    - Expose tables as MCP resources
    - Support Context7 token limits
    - Include caching layer
    - TypeScript with full type safety
  `,
  context7: {
    reference: '/supabase/supabase',
    patterns: ['mcp-best-practices', 'postgresql-integration'],
  },
})

// 생성된 코드:
// - src/index.ts (MCP server entry)
// - src/resources/ (resource handlers)
// - src/tools/ (MCP tools)
// - src/prompts/ (prompt templates)
// - tests/ (full test suite)
// - README.md (usage documentation)
```

### Industry Standards 준수

- **MCP Protocol**: ModelContextProtocol 사양 준수
- **Context7 Format**: 표준 응답 형식
- **OpenAPI 3.0**: REST API 문서화
- **TypeScript**: 완전한 타입 안전성
- **Security**: OAuth 2.0, API key 관리

### 코드 예제

```typescript
// 생성된 MCP 서버 구조
// src/index.ts
import { Server } from '@modelcontextprotocol/sdk/server/index.js'
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js'

const server = new Server(
  {
    name: 'postgresql-mcp',
    version: '1.0.0',
  },
  {
    capabilities: {
      resources: {},
      tools: {},
      prompts: {},
    },
  }
)

// Resource handler
server.setRequestHandler(ListResourcesRequestSchema, async () => {
  return {
    resources: [
      {
        uri: 'postgresql://tables/*',
        name: 'Database Tables',
        mimeType: 'application/json',
      },
    ],
  }
})

// Tool handler
server.setRequestHandler(CallToolRequestSchema, async (request) => {
  if (request.params.name === 'query') {
    const result = await executeQuery(request.params.arguments.sql)
    return {
      content: [
        {
          type: 'text',
          text: JSON.stringify(result),
        },
      ],
    }
  }
})

// Start server
const transport = new StdioServerTransport()
await server.connect(transport)
```

## Document Processing Skill

### 개요

Word, PDF, PowerPoint, Excel 파일을 AI 기반으로 처리하는 통합 문서 파이프라인입니다. 엔터프라이즈 워크플로우와 완벽하게 통합됩니다.

### 핵심 기능

**AI Content Extraction**:
```typescript
// Skill 호출로 문서 처리
Skill("moai-advanced-document-processing", {
  files: ['contract.pdf', 'report.docx', 'data.xlsx'],
  tasks: [
    'extract_text',
    'identify_entities',
    'summarize',
    'classify',
    'extract_tables',
  ],
  aiModel: 'claude-3-5-sonnet',
  outputFormat: 'structured-json',
})

// 생성된 결과:
{
  "contract.pdf": {
    "text": "...",
    "entities": {
      "parties": ["Company A", "Company B"],
      "dates": ["2024-01-01", "2025-01-01"],
      "amounts": ["$100,000"],
    },
    "summary": "This contract establishes...",
    "classification": "legal-contract",
    "metadata": {
      "pages": 15,
      "wordCount": 3500,
      "language": "en",
    }
  },
  "data.xlsx": {
    "tables": [
      {
        "name": "Sales Data",
        "rows": 1000,
        "columns": ["Date", "Product", "Amount", "Customer"],
        "preview": [...],
      }
    ],
    "summary": "Sales data for Q4 2024...",
  }
}
```

**Enterprise Workflows**:
```typescript
// 자동화된 문서 처리 파이프라인
import { DocumentPipeline } from '@moai/document-processing'

const pipeline = new DocumentPipeline({
  input: 's3://documents/inbox/',
  processors: [
    // 1. OCR for scanned documents
    {
      type: 'ocr',
      languages: ['en', 'ko'],
    },
    // 2. AI-based classification
    {
      type: 'classify',
      model: 'claude-3-5-sonnet',
      categories: ['invoice', 'contract', 'report', 'other'],
    },
    // 3. Entity extraction
    {
      type: 'extract-entities',
      entities: ['person', 'organization', 'date', 'money'],
    },
    // 4. Summarization
    {
      type: 'summarize',
      maxLength: 500,
    },
    // 5. Routing based on classification
    {
      type: 'route',
      rules: {
        invoice: 's3://documents/invoices/',
        contract: 's3://documents/contracts/',
        report: 's3://documents/reports/',
      },
    },
  ],
  output: {
    format: 'json',
    metadata: true,
    thumbnails: true,
  },
})

await pipeline.run()
```

### 지원 형식

| 형식 | 기능 | AI 처리 |
|------|------|---------|
| PDF | 텍스트, 이미지, 표 추출 | ✅ OCR, 요약, 분류 |
| DOCX | 텍스트, 서식, 표 추출 | ✅ 구조 분석, 편집 |
| PPTX | 슬라이드, 노트, 이미지 | ✅ 내용 요약 |
| XLSX | 시트, 표, 차트 | ✅ 데이터 분석 |

### 코드 예제

```typescript
// 고급 PDF 처리
import { PDFProcessor } from '@moai/document-processing'

const processor = new PDFProcessor({
  aiModel: 'claude-3-5-sonnet',
  features: {
    ocr: true,
    tables: true,
    images: true,
    forms: true,
  },
})

const result = await processor.process('contract.pdf', {
  tasks: [
    {
      type: 'extract-clauses',
      categories: ['payment-terms', 'termination', 'liability'],
    },
    {
      type: 'risk-analysis',
      focus: ['legal', 'financial'],
    },
    {
      type: 'compare',
      template: 'standard-contract.pdf',
      highlightDifferences: true,
    },
  ],
})

console.log(result.clauses)
console.log(result.risks)
console.log(result.differences)
```

## Artifacts Builder Skill

### 개요

모던 React 컴포넌트를 Tailwind CSS와 shadcn/ui를 활용하여 AI 기반으로 생성합니다. 디자인 시스템 준수 및 접근성 표준을 자동으로 적용합니다.

### 핵심 기능

**React + Tailwind + shadcn/ui**:
```typescript
// Skill 호출로 컴포넌트 생성
Skill("moai-advanced-artifacts-builder", {
  prompt: `
    Create a dashboard component with:
    - Responsive grid layout
    - Chart widgets (line, bar, pie)
    - Data table with sorting/filtering
    - Dark mode support
    - Accessibility (WCAG 2.1 AA)
  `,
  design: {
    system: 'shadcn-ui',
    theme: 'default',
    customColors: {
      primary: '#0F172A',
      accent: '#3B82F6',
    },
  },
})

// 생성된 컴포넌트:
// components/Dashboard.tsx
'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { LineChart, BarChart, PieChart } from '@/components/ui/charts'
import { DataTable } from '@/components/ui/data-table'

export function Dashboard() {
  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <Card>
        <CardHeader>
          <CardTitle>Total Revenue</CardTitle>
        </CardHeader>
        <CardContent>
          <LineChart data={revenueData} />
        </CardContent>
      </Card>
      {/* 더 많은 위젯들... */}
    </div>
  )
}
```

**AI-Powered Optimization**:
- **Performance**: Lazy loading, Code splitting 자동 적용
- **Accessibility**: ARIA labels, keyboard navigation 자동 추가
- **Responsive**: Mobile-first 디자인 자동 생성
- **Type Safety**: TypeScript props 자동 생성

### 코드 예제

```typescript
// 복잡한 폼 컴포넌트 생성
Skill("moai-advanced-artifacts-builder", {
  prompt: `
    Create a multi-step form for user onboarding:

    Step 1: Personal Information
    - Full name (required)
    - Email (validated)
    - Phone number (formatted)

    Step 2: Company Details
    - Company name
    - Industry (dropdown)
    - Size (radio buttons)

    Step 3: Preferences
    - Communication preferences (checkboxes)
    - Newsletter subscription (toggle)

    Features:
    - Form validation with zod
    - Progress indicator
    - Save draft functionality
    - Error handling
    - Success animation
  `,
})

// 생성된 코드는 즉시 사용 가능:
// - TypeScript 타입 정의
// - Zod 스키마
// - React Hook Form 통합
// - Tailwind 스타일링
// - shadcn/ui 컴포넌트
// - 접근성 최적화
```

## Internal Communications Skill

### 개요

엔터프라이즈 환경의 내부 커뮤니케이션을 AI로 최적화합니다. 이메일, 슬랙 메시지, 보고서 작성을 자동화하고 템플릿 라이브러리를 제공합니다.

### 핵심 기능

**AI Content Generation**:
```typescript
// 전문적인 이메일 자동 생성
Skill("moai-advanced-internal-comms", {
  type: 'email',
  context: {
    recipient: 'executive-team',
    subject: 'Q4 Performance Report',
    tone: 'professional',
    length: 'detailed',
  },
  data: {
    revenue: '$2.5M',
    growth: '+15%',
    challenges: ['Market competition', 'Supply chain'],
    opportunities: ['New markets', 'Product expansion'],
  },
})

// 생성된 이메일:
/**
 * Subject: Q4 2024 Performance Report - Executive Summary
 *
 * Dear Executive Team,
 *
 * I am pleased to present our Q4 2024 performance report, highlighting
 * significant achievements and strategic insights for the upcoming quarter.
 *
 * Key Highlights:
 * - Revenue reached $2.5M, representing a 15% increase YoY
 * - Successfully expanded into three new markets
 * - Launched two new product lines with strong initial adoption
 *
 * Challenges Addressed:
 * 1. Market Competition: Implemented competitive pricing strategy...
 * 2. Supply Chain: Diversified supplier base to ensure continuity...
 *
 * Strategic Opportunities:
 * - Market Expansion: Analysis shows potential in APAC region...
 * - Product Development: Customer feedback indicates demand for...
 *
 * [Detailed analysis attached]
 *
 * Best regards,
 * [Your Name]
 */
```

**Template Library**:
```typescript
// 재사용 가능한 템플릿
const templates = {
  weekly_update: {
    recipients: ['team'],
    sections: ['achievements', 'challenges', 'next_steps'],
    tone: 'casual',
  },
  project_proposal: {
    recipients: ['stakeholders'],
    sections: ['executive_summary', 'objectives', 'timeline', 'budget'],
    tone: 'professional',
  },
  incident_report: {
    recipients: ['management', 'technical_team'],
    sections: ['summary', 'timeline', 'impact', 'resolution', 'prevention'],
    tone: 'formal',
    urgency: 'high',
  },
}

// 템플릿 사용
Skill("moai-advanced-internal-comms", {
  template: 'incident_report',
  data: {
    incident: 'Database outage',
    duration: '45 minutes',
    affectedUsers: 1200,
    resolution: 'Failover to backup',
  },
})
```

### 지원 형식

- **Email**: 프로페셔널 이메일 작성
- **Slack/Teams**: 팀 메시지 최적화
- **Reports**: 주간/월간 리포트 생성
- **Presentations**: 슬라이드 콘텐츠 작성
- **Announcements**: 공지사항 초안

## Playwright Testing Skill

### 개요

웹 애플리케이션의 E2E 테스트를 Playwright로 자동화합니다. AI가 테스트 시나리오를 생성하고 유지보수합니다.

### 핵심 기능

**AI Test Generation**:
```typescript
// Skill 호출로 테스트 자동 생성
Skill("moai-advanced-playwright-testing", {
  url: 'https://myapp.com',
  scenarios: [
    'user-signup-flow',
    'checkout-process',
    'dashboard-interaction',
    'form-validation',
  ],
  coverage: {
    pages: ['/', '/signup', '/checkout', '/dashboard'],
    features: ['authentication', 'payment', 'data-entry'],
  },
})

// 생성된 테스트:
// tests/user-signup-flow.spec.ts
import { test, expect } from '@playwright/test'

test.describe('User Signup Flow', () => {
  test('should complete signup successfully', async ({ page }) => {
    // 1. Navigate to signup page
    await page.goto('https://myapp.com/signup')

    // 2. Fill in form
    await page.fill('[name="email"]', 'user@example.com')
    await page.fill('[name="password"]', 'SecurePass123!')
    await page.fill('[name="confirmPassword"]', 'SecurePass123!')

    // 3. Submit form
    await page.click('button[type="submit"]')

    // 4. Verify success
    await expect(page).toHaveURL(/.*dashboard/)
    await expect(page.locator('.welcome-message')).toContainText('Welcome')
  })

  test('should show validation errors', async ({ page }) => {
    await page.goto('https://myapp.com/signup')

    // Submit without filling
    await page.click('button[type="submit"]')

    // Verify errors
    await expect(page.locator('.error-email')).toBeVisible()
    await expect(page.locator('.error-password')).toBeVisible()
  })

  test('should handle duplicate email', async ({ page }) => {
    await page.goto('https://myapp.com/signup')

    await page.fill('[name="email"]', 'existing@example.com')
    await page.fill('[name="password"]', 'SecurePass123!')
    await page.click('button[type="submit"]')

    await expect(page.locator('.error-duplicate')).toBeVisible()
  })
})
```

**Visual Regression Testing**:
```typescript
// 스크린샷 비교 테스트 자동 생성
test('visual regression - homepage', async ({ page }) => {
  await page.goto('https://myapp.com')

  // Desktop
  await page.setViewportSize({ width: 1920, height: 1080 })
  await expect(page).toHaveScreenshot('homepage-desktop.png')

  // Mobile
  await page.setViewportSize({ width: 375, height: 667 })
  await expect(page).toHaveScreenshot('homepage-mobile.png')
})
```

### 코드 예제

```typescript
// 복잡한 사용자 플로우 테스트
Skill("moai-advanced-playwright-testing", {
  prompt: `
    Create comprehensive tests for e-commerce checkout:

    Flow:
    1. Browse products
    2. Add to cart
    3. Apply coupon code
    4. Enter shipping information
    5. Select payment method
    6. Review order
    7. Complete purchase
    8. Verify confirmation email

    Edge cases:
    - Invalid coupon
    - Out of stock items
    - Payment failure
    - Session timeout
  `,
})

// 생성된 테스트는 모든 엣지 케이스를 커버하며
// 자동으로 재시도 로직과 에러 핸들링 포함
```

## 나머지 Advanced Skills

### Development Chat

AI 페어 프로그래밍 어시스턴트:
- 실시간 코드 리뷰
- 리팩토링 제안
- 버그 디버깅 지원
- 아키텍처 토론

### Prompt Engineering

LLM 프롬프트 최적화:
- Few-shot 예제 생성
- Chain-of-Thought 프롬프팅
- 프롬프트 템플릿 라이브러리
- A/B 테스팅

### Video Analysis

영상 콘텐츠 분석:
- 자막 생성
- 키 프레임 추출
- 객체/장면 인식
- 요약 생성

## Best Practices

### 1. Skill 조합

```typescript
// 여러 Advanced Skills를 함께 사용
Skill("moai-advanced-document-processing", {
  files: ['requirements.pdf'],
})

// 추출된 요구사항으로 MCP 서버 생성
Skill("moai-advanced-mcp-builder", {
  requirements: extractedRequirements,
})

// 생성된 서버에 대한 테스트 작성
Skill("moai-advanced-playwright-testing", {
  mcpServer: generatedServer,
})
```

### 2. Context7 활용

```typescript
// Context7에서 최신 패턴 가져오기
Skill("moai-advanced-artifacts-builder", {
  context7: {
    reference: '/shadcn-ui/ui',
    patterns: ['dashboard', 'data-table', 'charts'],
  },
})
```

### 3. 반복 작업 자동화

```typescript
// 정기 리포트 자동 생성
Skill("moai-advanced-internal-comms", {
  schedule: 'weekly',
  template: 'weekly_update',
  dataSources: ['github', 'jira', 'analytics'],
})
```

## 다음 단계

- [Skill 개발 가이드](/ko/skills/skill-development) - 커스텀 Skill 만들기
- [Context7 통합](/ko/skills/context7-integration) - AI 기반 개발
- [Validation System](/ko/skills/validation-system) - 품질 보증
- [Foundation Skills](/ko/skills/foundation) - 기본 Skills 가이드

## 참고 자료

- [MoAI-ADK 공식 문서](https://moai-adk.dev)
- [Context7 가이드](https://context7.com/docs)
- [Playwright 문서](https://playwright.dev)
- [shadcn/ui](https://ui.shadcn.com)
