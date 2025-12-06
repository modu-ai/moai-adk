# MCP Integration Specialist Examples

실용적인 예시를 통해 MCP 통합 패턴과 멀티 서비스 오케스트레이션을 학습합니다.

---

## Example 1: Figma 디자인 토큰 추출

**Scenario**: Figma 파일에서 디자인 토큰을 추출하여 TypeScript 코드로 변환하는 상황

**Input**:
```python
from moai_integration_mcp import UniversalMCPServer

# MCP 서버 초기화
mcp_server = UniversalMCPServer("design-integration")
mcp_server.setup_connectors({
    'figma': {'api_key': os.getenv('FIGMA_TOKEN')}
})

# 디자인 토큰 동기화
tokens = await mcp_server.invoke_tool("sync_figma_tokens", {
    "file_id": "abc123XYZ",
    "output_format": "typescript",
    "include_variants": True
})
```

**Output**:
```typescript
// generated/design-tokens.ts
export const colors = {
  primary: {
    50: '#E3F2FD',
    100: '#BBDEFB',
    500: '#2196F3',
    900: '#0D47A1'
  },
  semantic: {
    success: '#4CAF50',
    warning: '#FF9800',
    error: '#F44336'
  }
} as const;

export const typography = {
  heading: {
    h1: { fontSize: '2.5rem', fontWeight: 700, lineHeight: 1.2 },
    h2: { fontSize: '2rem', fontWeight: 600, lineHeight: 1.3 }
  },
  body: {
    regular: { fontSize: '1rem', fontWeight: 400, lineHeight: 1.5 }
  }
} as const;

export const spacing = {
  xs: '0.25rem',
  sm: '0.5rem',
  md: '1rem',
  lg: '1.5rem',
  xl: '2rem'
} as const;
```

**Explanation**: Figma API를 통해 디자인 시스템의 색상, 타이포그래피, 간격 등의 토큰을 추출하고 TypeScript 상수로 변환합니다. include_variants 옵션으로 색상 변형도 함께 추출됩니다.

---

## Example 2: Notion 지식 베이스 분석

**Scenario**: Notion 데이터베이스에서 콘텐츠를 추출하고 AI로 분석하여 인사이트 도출

**Input**:
```python
# Notion 데이터베이스 쿼리
content = await mcp_server.invoke_tool("query_notion_database", {
    "database_id": "knowledge-base-db-id",
    "query": {
        "filter": {
            "property": "Status",
            "select": {"equals": "Published"}
        },
        "sorts": [
            {"property": "Created", "direction": "descending"}
        ]
    }
})

# AI 분석 실행
analysis = await mcp_server.invoke_tool("analyze_with_ai", {
    "content": json.dumps(content["results"]),
    "analysis_type": "best_practices",
    "include_key_points": True
})

# 구조화된 지식 베이스 생성
structured_kb = await mcp_server.invoke_tool("generate_ai_content", {
    "prompt": f"Create structured knowledge base from: {json.dumps(analysis)}",
    "max_tokens": 5000
})
```

**Output**:
```json
{
  "knowledge_base": {
    "title": "Development Best Practices",
    "categories": [
      {
        "name": "Code Quality",
        "practices": [
          {
            "title": "Test-Driven Development",
            "summary": "Write tests before implementation",
            "key_points": [
              "RED: Write failing test first",
              "GREEN: Implement minimum code to pass",
              "REFACTOR: Optimize without breaking tests"
            ],
            "source_count": 12
          }
        ]
      },
      {
        "name": "Security",
        "practices": [
          {
            "title": "Input Validation",
            "summary": "Always validate user input",
            "key_points": [
              "Use whitelist validation",
              "Sanitize before database operations",
              "Implement rate limiting"
            ],
            "source_count": 8
          }
        ]
      }
    ],
    "total_sources_analyzed": 47,
    "confidence_score": 0.92
  }
}
```

**Explanation**: Notion API로 게시된 문서를 조회하고, AI 분석을 통해 Best Practices를 추출합니다. 여러 문서에서 공통 패턴을 식별하여 구조화된 지식 베이스를 생성합니다.

---

## Example 3: Design-to-Code 파이프라인

**Scenario**: Figma 디자인을 React 컴포넌트로 자동 변환하는 전체 워크플로우

**Input**:
```python
async def complete_design_workflow(figma_file_id: str):
    """Figma 디자인에서 프로덕션 코드까지 완전 자동화"""

    # Phase 1: 디자인 데이터 추출
    design_data = await mcp_server.invoke_tool("extract_figma_components", {
        "file_id": figma_file_id,
        "include_tokens": True
    })

    # Phase 2: AI로 컴포넌트 스펙 분석
    component_specs = []
    for component in design_data["components"]:
        spec = await mcp_server.invoke_tool("analyze_with_ai", {
            "content": json.dumps(component),
            "analysis_type": "component_specification"
        })
        component_specs.append(spec)

    # Phase 3: React 컴포넌트 코드 생성
    generated_components = []
    for spec in component_specs:
        code = await mcp_server.invoke_tool("generate_ai_content", {
            "prompt": f"""
            Generate React component with TypeScript:
            - Component: {spec['name']}
            - Props: {spec['props']}
            - Styling: Use Tailwind CSS
            - Include accessibility attributes
            """,
            "max_tokens": 3000
        })
        generated_components.append(code)

    # Phase 4: Storybook 스토리 및 문서 생성
    documentation = await mcp_server.invoke_tool("generate_ai_content", {
        "prompt": f"Create Storybook stories and documentation for: {json.dumps(component_specs)}",
        "max_tokens": 4000
    })

    return {
        "components": generated_components,
        "documentation": documentation,
        "design_tokens": design_data["design_tokens"],
        "workflow_status": "completed"
    }
```

**Output**:
```typescript
// components/Button.tsx
import React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';

const buttonVariants = cva(
  'inline-flex items-center justify-center rounded-md font-medium transition-colors focus-visible:outline-none focus-visible:ring-2',
  {
    variants: {
      variant: {
        primary: 'bg-primary-500 text-white hover:bg-primary-600',
        secondary: 'bg-gray-100 text-gray-900 hover:bg-gray-200',
        outline: 'border border-gray-300 bg-transparent hover:bg-gray-50'
      },
      size: {
        sm: 'h-8 px-3 text-sm',
        md: 'h-10 px-4 text-base',
        lg: 'h-12 px-6 text-lg'
      }
    },
    defaultVariants: {
      variant: 'primary',
      size: 'md'
    }
  }
);

interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  isLoading?: boolean;
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, isLoading, children, ...props }, ref) => {
    return (
      <button
        ref={ref}
        className={buttonVariants({ variant, size, className })}
        disabled={isLoading}
        aria-busy={isLoading}
        {...props}
      >
        {isLoading ? <Spinner className="mr-2" /> : null}
        {children}
      </button>
    );
  }
);

Button.displayName = 'Button';
```

**Explanation**: 4단계 파이프라인으로 Figma 디자인을 프로덕션 React 코드로 변환합니다. 디자인 추출 → AI 스펙 분석 → 코드 생성 → 문서화 순서로 진행되며, Tailwind CSS와 접근성 속성이 자동으로 포함됩니다.

---

## Common Patterns

### Pattern 1: MCP 서버 초기화

다중 커넥터를 설정하는 표준 패턴입니다.

```python
from moai_integration_mcp import UniversalMCPServer, ServiceOrchestrator

# 서버 초기화
mcp_server = UniversalMCPServer("integration-server")

# 다중 커넥터 설정
mcp_server.setup_connectors({
    'figma': {
        'api_key': os.getenv('FIGMA_TOKEN'),
        'team_id': 'team-123'
    },
    'notion': {
        'api_key': os.getenv('NOTION_TOKEN'),
        'workspace_id': 'workspace-456'
    },
    'nano_banana': {
        'api_key': os.getenv('NANO_BANANA_TOKEN'),
        'model': 'claude-3-5-sonnet'
    }
})

# 오케스트레이터 등록
orchestrator = ServiceOrchestrator(mcp_server)
orchestrator.register_workflows()

# 서버 시작
mcp_server.start(port=3000)
```

### Pattern 2: 에러 핸들링과 재시도

서비스 호출 시 안정성을 보장하는 패턴입니다.

```python
from tenacity import retry, stop_after_attempt, wait_exponential

class ResilientMCPClient:
    def __init__(self, mcp_server):
        self.server = mcp_server

    @retry(
        stop=stop_after_attempt(3),
        wait=wait_exponential(multiplier=1, min=4, max=10)
    )
    async def invoke_with_retry(self, tool_name: str, params: dict):
        try:
            result = await self.server.invoke_tool(tool_name, params)
            return result
        except RateLimitError:
            # 속도 제한 시 대기 후 재시도
            await asyncio.sleep(60)
            raise
        except ServiceUnavailableError as e:
            # 서비스 불가 시 폴백 처리
            return self._handle_fallback(tool_name, params, e)

    def _handle_fallback(self, tool_name, params, error):
        return {
            "status": "fallback",
            "message": f"Service unavailable: {error}",
            "cached_result": self._get_cached_result(tool_name, params)
        }
```

### Pattern 3: 병렬 서비스 호출

독립적인 서비스 호출을 병렬로 실행하는 패턴입니다.

```python
async def parallel_service_calls():
    # 병렬 실행: 각 서비스 호출이 독립적
    results = await asyncio.gather(
        mcp_server.invoke_tool("extract_figma_components", {
            "file_id": "design-file"
        }),
        mcp_server.invoke_tool("query_notion_database", {
            "database_id": "docs-db"
        }),
        mcp_server.invoke_tool("analyze_with_ai", {
            "content": "existing data",
            "analysis_type": "summary"
        }),
        return_exceptions=True  # 개별 실패 허용
    )

    # 결과 처리
    figma_result, notion_result, ai_result = results

    # 실패한 호출 필터링
    successful_results = [r for r in results if not isinstance(r, Exception)]
    failed_results = [r for r in results if isinstance(r, Exception)]

    return {
        "successful": successful_results,
        "failed": len(failed_results),
        "partial_success": len(failed_results) > 0
    }
```

---

## Anti-Patterns (피해야 할 패턴)

### Anti-Pattern 1: API 키 하드코딩

**Problem**: 소스 코드에 API 키를 직접 작성

```python
# 잘못된 예시 - 보안 위험!
mcp_server.setup_connectors({
    'figma': {'api_key': 'figd_abc123...'}  # 절대 금지!
})
```

**Solution**: 환경 변수 또는 시크릿 매니저 사용

```python
# 올바른 예시
import os
from secret_manager import get_secret

mcp_server.setup_connectors({
    'figma': {'api_key': os.getenv('FIGMA_TOKEN')},
    # 또는 시크릿 매니저 사용
    'notion': {'api_key': get_secret('notion-api-key')}
})
```

### Anti-Pattern 2: 순차적 API 호출

**Problem**: 독립적인 API 호출을 순차적으로 실행하여 시간 낭비

```python
# 잘못된 예시 - 비효율적
figma_data = await mcp_server.invoke_tool("extract_figma_components", {...})
notion_data = await mcp_server.invoke_tool("query_notion_database", {...})
ai_analysis = await mcp_server.invoke_tool("analyze_with_ai", {...})
# 총 시간: 5초 + 3초 + 2초 = 10초
```

**Solution**: 독립적인 호출은 병렬 실행

```python
# 올바른 예시 - 효율적
figma_data, notion_data, ai_analysis = await asyncio.gather(
    mcp_server.invoke_tool("extract_figma_components", {...}),
    mcp_server.invoke_tool("query_notion_database", {...}),
    mcp_server.invoke_tool("analyze_with_ai", {...})
)
# 총 시간: max(5초, 3초, 2초) = 5초
```

### Anti-Pattern 3: 에러 핸들링 누락

**Problem**: API 실패 시 전체 워크플로우 중단

```python
# 잘못된 예시 - 에러 시 중단
async def fragile_workflow():
    result = await mcp_server.invoke_tool("some_api", {...})
    # API 실패 시 전체 중단
    return process(result)
```

**Solution**: 적절한 에러 핸들링과 폴백 로직

```python
# 올바른 예시 - 복원력 있는 처리
async def resilient_workflow():
    try:
        result = await mcp_server.invoke_tool("some_api", {...})
        return process(result)
    except ServiceUnavailableError:
        # 캐시된 결과 사용
        return get_cached_result()
    except RateLimitError:
        # 재시도 스케줄링
        return schedule_retry()
    except Exception as e:
        # 로깅 후 부분 결과 반환
        logger.error(f"API call failed: {e}")
        return partial_result()
```

### Anti-Pattern 4: 무제한 토큰 요청

**Problem**: AI 생성 시 토큰 제한 없이 요청하여 비용 폭증

```python
# 잘못된 예시 - 비용 위험
await mcp_server.invoke_tool("generate_ai_content", {
    "prompt": "Write comprehensive documentation...",
    # max_tokens 미지정 → 기본값이 매우 클 수 있음
})
```

**Solution**: 항상 적절한 토큰 제한 설정

```python
# 올바른 예시 - 비용 통제
await mcp_server.invoke_tool("generate_ai_content", {
    "prompt": "Write comprehensive documentation...",
    "max_tokens": 4000,  # 명시적 제한
    "temperature": 0.7   # 일관성을 위한 온도 설정
})
```

---

## Integration Checklist

MCP 통합 시 확인해야 할 항목:

| 항목 | 확인 |
|------|------|
| API 키는 환경 변수에 저장했는가? | |
| 재시도 로직을 구현했는가? | |
| 속도 제한 처리를 추가했는가? | |
| 에러 핸들링이 완비되었는가? | |
| 토큰/비용 제한을 설정했는가? | |
| 병렬 실행 가능한 부분을 식별했는가? | |
| 타임아웃을 설정했는가? | |
| 로깅을 구현했는가? | |

---

Version: 1.0.0
Last Updated: 2025-12-06
