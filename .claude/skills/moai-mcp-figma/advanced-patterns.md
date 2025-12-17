# Figma MCP Advanced Patterns

완전한 Design-to-Code 파이프라인, 컴포넌트 변형 처리, Figma Variables 변환, 실시간 동기화, 디자인 시스템 검증을 위한 고급 패턴 가이드입니다.

---

## Complete Design-to-Code Pipeline

### 전체 워크플로우 자동화

```python
async def complete_design_to_code_pipeline(
    figma_file_id: str,
    target_framework: str = "react",
    target_library: str = "shadcn"
):
    """Figma에서 프로덕션 코드까지 완전 자동화 파이프라인."""

    # Phase 1: 디자인 토큰 추출 및 동기화
    design_tokens = await figma_server.invoke_tool("sync_figma_tokens", {
        "file_id": figma_file_id,
        "token_types": ["colors", "typography", "spacing", "effects"],
        "output_format": "typescript"
    })

    # Phase 2: 컴포넌트 메타데이터 추출
    components = await figma_server.invoke_tool("extract_figma_components", {
        "file_id": figma_file_id,
        "include_variants": True,
        "include_instances": True
    })

    # Phase 3: 각 컴포넌트를 코드로 변환
    generated_components = []
    for component in components["components"]:
        code = await figma_server.invoke_tool("generate_react_component", {
            "file_id": figma_file_id,
            "component_id": component["id"],
            "target_library": target_library,
            "include_typescript": True,
            "include_stories": True,
            "include_tests": True
        })
        generated_components.append(code)

    # Phase 4: 자산 추출 및 최적화
    assets = await figma_server.invoke_tool("export_all_assets", {
        "file_id": figma_file_id,
        "optimize": True,
        "output_directory": "./src/assets"
    })

    # Phase 5: 디자인 시스템 문서 생성
    documentation = await figma_server.invoke_tool("generate_design_system_docs", {
        "file_id": figma_file_id,
        "components": components,
        "tokens": design_tokens,
        "include_examples": True
    })

    return {
        "design_tokens": design_tokens,
        "components": generated_components,
        "assets": assets,
        "documentation": documentation,
        "pipeline_status": "completed"
    }
```

### 병렬 처리를 통한 성능 최적화

```python
import asyncio

async def optimized_pipeline(figma_file_id: str):
    """병렬 처리를 통한 성능 최적화 파이프라인."""

    # 병렬로 독립적인 작업 실행
    tokens_task = figma_server.invoke_tool("sync_figma_tokens", {
        "file_id": figma_file_id,
        "token_types": ["colors", "typography", "spacing"]
    })

    components_task = figma_server.invoke_tool("extract_figma_components", {
        "file_id": figma_file_id,
        "include_variants": True
    })

    assets_task = figma_server.invoke_tool("export_all_assets", {
        "file_id": figma_file_id,
        "optimize": True
    })

    # 모든 작업 동시 실행
    tokens, components, assets = await asyncio.gather(
        tokens_task,
        components_task,
        assets_task
    )

    # 순차적으로 컴포넌트 생성 (토큰 정보 필요)
    generated = []
    for comp in components["components"]:
        code = await figma_server.invoke_tool("generate_react_component", {
            "file_id": figma_file_id,
            "component_id": comp["id"],
            "design_tokens": tokens  # 추출된 토큰 사용
        })
        generated.append(code)

    return {
        "tokens": tokens,
        "components": generated,
        "assets": assets
    }
```

---

## Component Variants Processing

### Component Set에서 모든 변형 추출

```python
# Button Component Set의 모든 변형 추출
button_variants = await figma_server.invoke_tool("extract_component_variants", {
    "file_id": file_id,
    "component_set_id": "123:456",  # Component Set ID
    "generate_union_types": True
})

# TypeScript 유니온 타입 생성:
# export type ButtonVariant = 'primary' | 'secondary' | 'outline';
# export type ButtonSize = 'sm' | 'md' | 'lg';
```

### 변형 조합을 Props로 변환

```python
# 모든 변형 조합을 컴포넌트 Props로 변환
component_with_variants = await figma_server.invoke_tool("generate_component_with_variants", {
    "file_id": file_id,
    "component_set_id": "123:456",
    "target_framework": "react",
    "prop_mapping": {
        "Variant": "variant",
        "Size": "size",
        "State": "state"
    }
})

# 생성된 컴포넌트:
# export interface ButtonProps {
#   variant?: 'primary' | 'secondary' | 'outline';
#   size?: 'sm' | 'md' | 'lg';
#   state?: 'default' | 'hover' | 'disabled';
# }
```

### 변형별 스타일 자동 생성

```python
async def generate_variant_styles(component_set_id: str):
    """Component Set의 모든 변형에 대한 스타일 생성."""

    # 변형 데이터 추출
    variants = await figma_server.invoke_tool("extract_component_variants", {
        "file_id": file_id,
        "component_set_id": component_set_id,
        "include_styles": True
    })

    # Tailwind CSS 변형 생성
    tailwind_variants = await figma_server.invoke_tool("generate_tailwind_variants", {
        "variants": variants,
        "output_format": "cva"  # class-variance-authority
    })

    # 생성된 CVA 설정:
    # const buttonVariants = cva(
    #   "base-button-classes",
    #   {
    #     variants: {
    #       variant: {
    #         primary: "bg-blue-500 text-white",
    #         secondary: "bg-gray-500 text-white",
    #         outline: "border border-gray-300"
    #       },
    #       size: {
    #         sm: "px-3 py-1.5 text-sm",
    #         md: "px-4 py-2 text-base",
    #         lg: "px-6 py-3 text-lg"
    #       }
    #     }
    #   }
    # )

    return tailwind_variants
```

---

## Figma Variables to Design Tokens

### Variables 추출 및 변환

```python
# Figma Variables (색상, 숫자, 문자열 등) 추출
variables = await figma_server.invoke_tool("extract_figma_variables", {
    "file_id": file_id,
    "variable_collections": ["Colors", "Spacing", "Typography"],
    "include_modes": True  # Light/Dark 모드 포함
})

# 결과 구조:
# {
#   "Colors": {
#     "modes": ["Light", "Dark"],
#     "variables": {
#       "primary": {
#         "Light": "#3b82f6",
#         "Dark": "#60a5fa"
#       }
#     }
#   }
# }
```

### 테마별 토큰 생성

```python
# Light/Dark 테마별 CSS 변수 생성
themed_tokens = await figma_server.invoke_tool("generate_themed_tokens", {
    "variables": variables,
    "output_format": "css",
    "theme_selector": "data-theme"
})

# 생성된 CSS:
# [data-theme="light"] {
#   --color-primary: #3b82f6;
# }
# [data-theme="dark"] {
#   --color-primary: #60a5fa;
# }
```

### 시맨틱 토큰 시스템 구축

```python
async def build_semantic_token_system(file_id: str):
    """Figma Variables에서 시맨틱 토큰 시스템 구축."""

    # 프리미티브 토큰 추출 (기본 색상, 간격 등)
    primitive_tokens = await figma_server.invoke_tool("extract_figma_variables", {
        "file_id": file_id,
        "variable_collections": ["Primitives"],
        "token_type": "primitive"
    })

    # 시맨틱 토큰 추출 (버튼 색상, 텍스트 색상 등)
    semantic_tokens = await figma_server.invoke_tool("extract_figma_variables", {
        "file_id": file_id,
        "variable_collections": ["Semantic"],
        "token_type": "semantic",
        "reference_primitives": primitive_tokens
    })

    # 계층적 토큰 시스템 생성
    token_system = await figma_server.invoke_tool("generate_token_system", {
        "primitives": primitive_tokens,
        "semantics": semantic_tokens,
        "output_formats": ["css", "scss", "typescript", "json"]
    })

    return token_system
```

---

## Real-time Sync and Webhooks

### Figma 파일 변경 감지

```python
# 파일 버전 추적
version_history = await figma_server.invoke_tool("get_file_versions", {
    "file_id": file_id
})

# 최신 변경 사항 확인
latest_changes = await figma_server.invoke_tool("check_file_changes", {
    "file_id": file_id,
    "since_version": "last_sync_version"
})

# 변경된 컴포넌트만 재생성
if latest_changes["has_changes"]:
    updated_components = await figma_server.invoke_tool("sync_changed_components", {
        "file_id": file_id,
        "changed_nodes": latest_changes["changed_nodes"]
    })
```

### 자동 동기화 워크플로우

```python
import asyncio
from datetime import datetime

async def auto_sync_workflow(file_id: str, check_interval: int = 300):
    """자동 동기화 워크플로우 (5분마다 체크)."""

    last_version = None

    while True:
        try:
            # 최신 버전 확인
            current_version = await figma_server.invoke_tool("get_current_version", {
                "file_id": file_id
            })

            # 버전 변경 감지
            if last_version and current_version != last_version:
                print(f"[{datetime.now()}] New version detected: {current_version}")

                # 변경 사항 추출
                changes = await figma_server.invoke_tool("get_version_diff", {
                    "file_id": file_id,
                    "from_version": last_version,
                    "to_version": current_version
                })

                # 변경된 컴포넌트 재생성
                if changes["components"]:
                    await regenerate_components(file_id, changes["components"])

                # 변경된 토큰 재동기화
                if changes["styles"]:
                    await resync_design_tokens(file_id)

            last_version = current_version

        except Exception as e:
            print(f"Sync error: {e}")

        # 다음 체크까지 대기
        await asyncio.sleep(check_interval)
```

### GitHub Actions 통합

```yaml
# .github/workflows/figma-sync.yml
name: Figma Design Sync

on:
  schedule:
    - cron: '0 */6 * * *'  # 6시간마다 실행
  workflow_dispatch:  # 수동 실행 가능

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.11'

      - name: Install dependencies
        run: pip install moai-mcp-figma

      - name: Sync Figma Design System
        env:
          FIGMA_TOKEN: ${{ secrets.FIGMA_TOKEN }}
        run: python scripts/sync-figma.py

      - name: Create Pull Request
        uses: peter-evans/create-pull-request@v5
        with:
          title: 'chore: Sync Figma design system'
          body: 'Automated Figma design system synchronization'
          branch: figma-sync-${{ github.run_id }}
```

---

## Design System Validation

### 디자인 시스템 일관성 검사

```python
# 디자인 토큰 일관성 검증
validation = await figma_server.invoke_tool("validate_design_system", {
    "file_id": file_id,
    "rules": [
        "consistent_spacing",
        "consistent_colors",
        "component_naming",
        "variant_completeness"
    ]
})

# 검증 결과:
# {
#   "errors": [
#     {
#       "rule": "consistent_spacing",
#       "message": "Non-standard spacing found: 13px",
#       "node_id": "123:456"
#     }
#   ],
#   "warnings": [...],
#   "passed": false
# }
```

### 커스텀 검증 규칙

```python
async def validate_with_custom_rules(file_id: str):
    """커스텀 검증 규칙을 사용한 디자인 시스템 검증."""

    custom_rules = {
        "color_contrast": {
            "type": "accessibility",
            "min_ratio": 4.5,  # WCAG AA 기준
            "check_text_on_background": True
        },
        "spacing_scale": {
            "type": "consistency",
            "allowed_values": [4, 8, 12, 16, 24, 32, 48, 64],
            "tolerance": 0  # 정확히 일치해야 함
        },
        "component_naming": {
            "type": "convention",
            "pattern": r"^[A-Z][a-zA-Z]+(/[A-Z][a-zA-Z]+)*$",
            "examples": ["Button/Primary", "Card/Header"]
        }
    }

    validation_result = await figma_server.invoke_tool("validate_with_rules", {
        "file_id": file_id,
        "custom_rules": custom_rules,
        "generate_report": True
    })

    return validation_result
```

---

## Advanced Component Generation Strategies

### 복합 컴포넌트 생성

```python
# 중첩된 컴포넌트 구조 생성
complex_component = await figma_server.invoke_tool("generate_complex_component", {
    "file_id": file_id,
    "component_id": "123:456",
    "resolve_nested": True,  # 중첩 컴포넌트 자동 해결
    "extract_subcomponents": True,  # 하위 컴포넌트 별도 추출
    "target_framework": "react"
})

# 생성된 구조:
# components/
# ├── Card.tsx (메인 컴포넌트)
# ├── CardHeader.tsx (서브 컴포넌트)
# ├── CardContent.tsx (서브 컴포넌트)
# └── CardFooter.tsx (서브 컴포넌트)
```

### 반응형 컴포넌트 생성

```python
# Figma의 반응형 제약을 CSS로 변환
responsive_component = await figma_server.invoke_tool("generate_responsive_component", {
    "file_id": file_id,
    "component_id": "123:456",
    "breakpoints": {
        "mobile": 375,
        "tablet": 768,
        "desktop": 1440
    },
    "responsive_strategy": "container-queries"  # container-queries 또는 media-queries
})
```

### AI 기반 컴포넌트 최적화

```python
async def ai_optimized_component_generation(file_id: str, component_id: str):
    """AI를 활용한 컴포넌트 생성 및 최적화."""

    # 1. 기본 컴포넌트 생성
    base_component = await figma_server.invoke_tool("generate_react_component", {
        "file_id": file_id,
        "component_id": component_id
    })

    # 2. AI를 통한 접근성 개선
    accessible_component = await figma_server.invoke_tool("enhance_accessibility", {
        "component_code": base_component,
        "wcag_level": "AA",
        "add_aria_labels": True,
        "keyboard_navigation": True
    })

    # 3. AI를 통한 성능 최적화
    optimized_component = await figma_server.invoke_tool("optimize_performance", {
        "component_code": accessible_component,
        "lazy_loading": True,
        "code_splitting": True,
        "bundle_size_target": "10kb"
    })

    # 4. AI를 통한 테스트 생성
    with_tests = await figma_server.invoke_tool("generate_tests", {
        "component_code": optimized_component,
        "test_framework": "vitest",
        "coverage_target": 90
    })

    return with_tests
```

---

## Performance Optimization

### 캐싱 전략

```python
from functools import lru_cache
import hashlib

class FigmaCacheManager:
    def __init__(self):
        self.cache = {}

    def get_cache_key(self, file_id: str, node_id: str, version: str):
        """캐시 키 생성."""
        key = f"{file_id}:{node_id}:{version}"
        return hashlib.sha256(key.encode()).hexdigest()

    async def get_or_fetch_component(self, file_id: str, node_id: str):
        """캐시에서 컴포넌트 가져오기 또는 Figma에서 추출."""

        # 현재 버전 확인
        version = await figma_server.invoke_tool("get_current_version", {
            "file_id": file_id
        })

        cache_key = self.get_cache_key(file_id, node_id, version)

        # 캐시 확인
        if cache_key in self.cache:
            return self.cache[cache_key]

        # Figma에서 추출
        component = await figma_server.invoke_tool("extract_component", {
            "file_id": file_id,
            "node_id": node_id
        })

        # 캐시 저장
        self.cache[cache_key] = component

        return component
```

### 배치 처리

```python
async def batch_component_generation(file_id: str, component_ids: list):
    """여러 컴포넌트를 배치로 처리."""

    # 컴포넌트를 그룹으로 나누기 (한 번에 10개씩)
    batch_size = 10
    batches = [component_ids[i:i+batch_size] for i in range(0, len(component_ids), batch_size)]

    all_components = []

    for batch in batches:
        # 각 배치를 병렬로 처리
        tasks = [
            figma_server.invoke_tool("generate_react_component", {
                "file_id": file_id,
                "component_id": comp_id
            })
            for comp_id in batch
        ]

        batch_results = await asyncio.gather(*tasks)
        all_components.extend(batch_results)

    return all_components
```

---

*이 문서는 Figma MCP 통합의 고급 패턴을 다룹니다. 기본 사용법은 SKILL.md를 참조하세요.*
